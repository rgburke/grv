package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
)

const (
	// GitRepositoryDirectoryName is the name of the git directory in a git repository
	GitRepositoryDirectoryName = ".git"
	updatedRefChannelSize      = 256
)

// OnCommitsLoaded is called when all commits are loaded for the specified oid
type OnCommitsLoaded func(Ref) error

// OnRefsLoaded is called when all refs have been loaded and processed
type OnRefsLoaded func([]Ref) error

// StatusListener is notified when git status has changed
type StatusListener interface {
	OnStatusChanged(status *Status)
}

// UpdatedRef contains the old and new Oid a ref points to
type UpdatedRef struct {
	OldRef Ref
	NewRef Ref
}

// String returns a string representation of an updated ref
func (updatedRef UpdatedRef) String() string {
	return fmt.Sprintf("%v: %v -> %v",
		updatedRef.NewRef.Name(), updatedRef.OldRef.Oid(), updatedRef.NewRef.Oid())
}

// RefStateListener is updated when changes to refs are detected
type RefStateListener interface {
	OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef)
}

// RepoData houses all data loaded from the repository
type RepoData interface {
	Path() string
	LoadHead() error
	LoadRefs(OnRefsLoaded)
	LoadCommits(Ref, OnCommitsLoaded) error
	Head() Ref
	Branches() (localBranches, remoteBranches []*Branch, loading bool)
	LocalTags() (tags []*Tag, loading bool)
	RefsForCommit(*Commit) *CommitRefs
	CommitSetState(Ref) CommitSetState
	Commits(ref Ref, startIndex, count uint) (<-chan *Commit, error)
	CommitByIndex(ref Ref, index uint) (*Commit, error)
	Commit(oid *Oid) (*Commit, error)
	AddCommitFilter(Ref, *CommitFilter) error
	RemoveCommitFilter(Ref) error
	DiffCommit(commit *Commit) (*Diff, error)
	DiffFile(statusType StatusType, path string) (*Diff, error)
	LoadStatus() (err error)
	RegisterStatusListener(StatusListener)
	RegisterRefStateListener(RefStateListener)
}

type commitSet interface {
	AddCommit(commit *Commit) (err error)
	Commit(index uint) (commit *Commit)
	CommitStream() <-chan *Commit
	SetLoading(loading bool)
	CommitSetState() CommitSetState
	Update(commonAncestor *Commit, update []*Commit)
	Clone() commitSet
}

type filteredCommitSet struct {
	commits      []*Commit
	loading      bool
	child        commitSet
	commitFilter *CommitFilter
	lock         sync.Mutex
}

func newBaseFilteredCommitSet() *filteredCommitSet {
	return newFilteredCommitSet(nil, nil)
}

func newFilteredCommitSet(child commitSet, commitFilter *CommitFilter) *filteredCommitSet {
	return &filteredCommitSet{
		commits:      make([]*Commit, 0),
		child:        child,
		commitFilter: commitFilter,
	}
}

func (filteredCommitSet *filteredCommitSet) initialiseFromCommitSet() {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if filteredCommitSet.hasChild() {
		for commit := range filteredCommitSet.child.CommitStream() {
			filteredCommitSet.addCommitIfFilterMatches(commit)
		}
	}
}

// CommitSet returns the child commit set of this filter
func (filteredCommitSet *filteredCommitSet) Child() commitSet {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	return filteredCommitSet.child
}

// HasChild returns true if this commitSet has a child commitSet
func (filteredCommitSet *filteredCommitSet) HasChild() bool {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	return filteredCommitSet.hasChild()
}

func (filteredCommitSet *filteredCommitSet) hasChild() bool {
	return !(filteredCommitSet.child == nil || reflect.ValueOf(filteredCommitSet.child).IsNil())
}

// AddCommit adds the commit to the child and then itself if the filter matches
func (filteredCommitSet *filteredCommitSet) AddCommit(commit *Commit) (err error) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if filteredCommitSet.hasChild() {
		if err = filteredCommitSet.child.AddCommit(commit); err != nil {
			return
		}

		filteredCommitSet.addCommitIfFilterMatches(commit)
	} else if filteredCommitSet.loading {
		filteredCommitSet.commits = append(filteredCommitSet.commits, commit)
	} else {
		err = fmt.Errorf("Cannot add commit when CommitSet is not in loading state")
	}

	return
}

func (filteredCommitSet *filteredCommitSet) addCommitIfFilterMatches(commit *Commit) {
	if filteredCommitSet.commitFilter.MatchesFilter(commit) {
		filteredCommitSet.commits = append(filteredCommitSet.commits, commit)
	}
}

// Commit returns the commit at the specified index
func (filteredCommitSet *filteredCommitSet) Commit(index uint) (commit *Commit) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if index < uint(len(filteredCommitSet.commits)) {
		commit = filteredCommitSet.commits[index]
	}

	return
}

// CommitStream returns a channel through which all the commits in this set can be read
func (filteredCommitSet *filteredCommitSet) CommitStream() <-chan *Commit {
	ch := make(chan *Commit)

	go func() {
		defer close(ch)
		var commit *Commit
		index := 0

		for {
			filteredCommitSet.lock.Lock()

			length := len(filteredCommitSet.commits)
			if index < length {
				commit = filteredCommitSet.commits[index]
			}

			filteredCommitSet.lock.Unlock()

			if commit != nil {
				ch <- commit
				commit = nil
				index++
			} else {
				return
			}
		}
	}()

	return ch
}

// SetLoading is defered onto the underlying raw commit set
func (filteredCommitSet *filteredCommitSet) SetLoading(loading bool) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if filteredCommitSet.hasChild() {
		filteredCommitSet.child.SetLoading(loading)
	} else {
		filteredCommitSet.loading = loading
	}
}

// CommitSetState returns the state of this commit set
func (filteredCommitSet *filteredCommitSet) CommitSetState() CommitSetState {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if filteredCommitSet.hasChild() {
		commitSetState := filteredCommitSet.child.CommitSetState()

		if commitSetState.filterState == nil {
			commitSetState.filterState = &CommitSetFilterState{
				unfilteredCommitNum: commitSetState.commitNum,
			}
		}

		commitSetState.commitNum = uint(len(filteredCommitSet.commits))
		commitSetState.filterState.filtersApplied++

		return commitSetState
	}

	return CommitSetState{
		loading:   filteredCommitSet.loading,
		commitNum: uint(len(filteredCommitSet.commits)),
	}
}

func (filteredCommitSet *filteredCommitSet) Update(commonAncestor *Commit, update []*Commit) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if filteredCommitSet.hasChild() {
		filteredCommitSet.child.Update(commonAncestor, update)
	}

	ancestorIndex := filteredCommitSet.commitIndex(commonAncestor)
	log.Debugf("Common ancestor index: %v", ancestorIndex)

	if ancestorIndex < 0 {
		ancestorIndex = -(ancestorIndex + 1)
	}

	filteredCommitSet.commits = append(update, filteredCommitSet.commits[ancestorIndex:]...)
}

func (filteredCommitSet *filteredCommitSet) commitIndex(commit *Commit) int {
	commits := filteredCommitSet.commits

	// Heuristic check
	// The majority of the time branches are updated simply
	// by fast forwarding to the latest commit. The common ancestor
	// will be the first commit in the array in this case.
	// Therefore do a quick check to see if the common ancestor
	// appears within the first 5 commits
	for i := 0; i < len(commits) && i < 5; i++ {
		if commits[i].oid.Equal(commit.oid) {
			return i
		}
	}

	var low, high, mid int
	targetDate := commit.commit.Author().When
	high = len(commits) - 1
	low = 0

	for low <= high {
		mid = (low + high) / 2
		date := commits[mid].commit.Author().When

		if targetDate.Before(date) {
			high = mid - 1
		} else if targetDate.After(date) {
			low = mid + 1
		} else {
			break
		}
	}

	if low > high {
		return -(low + 1)
	}

	if commits[mid].oid.Equal(commit.oid) {
		return mid
	}

	for i := 1; ; i++ {
		lowIndex := mid - i
		highIndex := mid + i

		if lowIndex > -1 {
			if commits[lowIndex].commit.Author().When.Equal(targetDate) {
				if commits[lowIndex].oid.Equal(commit.oid) {
					return lowIndex
				}
			} else {
				lowIndex = -1
			}
		}

		if highIndex < len(commits) {
			if commits[highIndex].commit.Author().When.Equal(targetDate) {
				if commits[highIndex].oid.Equal(commit.oid) {
					return highIndex
				}
			} else {
				highIndex = len(commits)
			}
		}

		if lowIndex < 0 && highIndex >= len(commits) {
			break
		}
	}

	return -(mid + 1)
}

func (filteredCommitSet *filteredCommitSet) Clone() commitSet {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	var child commitSet
	if filteredCommitSet.hasChild() {
		child = filteredCommitSet.child.Clone()
	}

	clone := newBaseFilteredCommitSet()
	clone.child = child
	clone.commitFilter = filteredCommitSet.commitFilter
	clone.commits = append([]*Commit(nil), filteredCommitSet.commits...)
	clone.loading = filteredCommitSet.loading

	return clone
}

// CommitSetState describes the current state of a commit set for a ref
type CommitSetState struct {
	loading     bool
	commitNum   uint
	filterState *CommitSetFilterState
}

// CommitSetFilterState describes filter information for a commit set
type CommitSetFilterState struct {
	unfilteredCommitNum uint
	filtersApplied      uint
}

type refSet struct {
	refs               map[string]Ref
	localBranchesList  []*Branch
	remoteBranchesList []*Branch
	tagsList           []*Tag
	loading            bool
	refStateListeners  []RefStateListener
	lock               sync.Mutex
}

func newRefSet() *refSet {
	return &refSet{
		refs: make(map[string]Ref),
	}
}

func (refSet *refSet) registerRefStateListener(refStateListener RefStateListener) {
	if refStateListener == nil {
		return
	}

	log.Debugf("Registering ref state listener %T", refStateListener)

	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	refSet.refStateListeners = append(refSet.refStateListeners, refStateListener)
}

func (refSet *refSet) startUpdate() bool {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	if refSet.loading {
		return false
	}

	refSet.loading = true

	return true
}

func (refSet *refSet) endUpdate() {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	refSet.loading = false
}

func (refSet *refSet) update(refs []Ref) (err error) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	log.Debugf("Updating refs in refSet")

	if !refSet.loading {
		return fmt.Errorf("RefSet not in loading state")
	}

	var addedRefs, removedRefs []Ref
	var updatedRefs []*UpdatedRef

	refMap := make(map[string]Ref)
	var localBranches, remoteBranches []*Branch
	var tags []*Tag

	for _, ref := range refs {
		existingRef, isExisting := refSet.refs[ref.Name()]

		if isExisting {
			if !existingRef.Oid().Equal(ref.Oid()) {
				updatedRefs = append(updatedRefs, &UpdatedRef{
					OldRef: existingRef,
					NewRef: ref,
				})
			}
		} else {
			addedRefs = append(addedRefs, ref)
		}

		refMap[ref.Name()] = ref

		switch rawRef := ref.(type) {
		case *Branch:
			if rawRef.IsRemote() {
				remoteBranches = append(remoteBranches, rawRef)
			} else {
				localBranches = append(localBranches, rawRef)
			}
		case *Tag:
			tags = append(tags, rawRef)
		}
	}

	for name, ref := range refSet.refs {
		_, stillExists := refMap[name]

		if !stillExists {
			removedRefs = append(removedRefs, ref)
		}
	}

	slice.Sort(localBranches, func(i, j int) bool {
		return localBranches[i].Name() < localBranches[j].Name()
	})
	slice.Sort(remoteBranches, func(i, j int) bool {
		return remoteBranches[i].Name() < remoteBranches[j].Name()
	})
	slice.Sort(tags, func(i, j int) bool {
		return tags[i].Name() < tags[j].Name()
	})

	slice.Sort(addedRefs, func(i, j int) bool {
		return addedRefs[i].Name() < addedRefs[j].Name()
	})
	slice.Sort(removedRefs, func(i, j int) bool {
		return removedRefs[i].Name() < removedRefs[j].Name()
	})
	slice.Sort(updatedRefs, func(i, j int) bool {
		return updatedRefs[i].NewRef.Name() < updatedRefs[j].NewRef.Name()
	})

	refSet.refs = refMap
	refSet.localBranchesList = localBranches
	refSet.remoteBranchesList = remoteBranches
	refSet.tagsList = tags

	if len(addedRefs) > 0 || len(removedRefs) > 0 || len(updatedRefs) > 0 {
		log.Debugf("Refs Changed - new: %v, removed: %v, updated: %v",
			len(addedRefs), len(removedRefs), len(updatedRefs))
		refSet.notifyRefStateListeners(addedRefs, removedRefs, updatedRefs)
	} else {
		log.Debugf("No new, removed or modified refs")
	}

	return
}

func (refSet *refSet) notifyRefStateListeners(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	refStateListeners := append([]RefStateListener(nil), refSet.refStateListeners...)

	go func() {
		for _, refStateListener := range refStateListeners {
			refStateListener.OnRefsChanged(addedRefs, removedRefs, updatedRefs)
		}
	}()
}

// CommitRefs contain all refs to a commit
type CommitRefs struct {
	tags     []*Tag
	branches []*Branch
}

type commitRefSet struct {
	commitRefs map[*Oid]*CommitRefs
	lock       sync.Mutex
}

func newCommitRefSet() *commitRefSet {
	commitRefSet := &commitRefSet{}
	commitRefSet.clear()
	return commitRefSet
}

func (commitRefSet *commitRefSet) clear() {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefSet.commitRefs = make(map[*Oid]*CommitRefs)
}

func (commitRefSet *commitRefSet) addTagForCommit(commit *Commit, newTag *Tag) {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefs, ok := commitRefSet.commitRefs[commit.oid]
	if !ok {
		commitRefs = &CommitRefs{}
		commitRefSet.commitRefs[commit.oid] = commitRefs
	}

	for _, tag := range commitRefs.tags {
		if tag.name == newTag.name {
			return
		}
	}

	commitRefs.tags = append(commitRefs.tags, newTag)
}

func (commitRefSet *commitRefSet) addBranchForCommit(commit *Commit, newBranch *Branch) {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefs, ok := commitRefSet.commitRefs[commit.oid]
	if !ok {
		commitRefs = &CommitRefs{}
		commitRefSet.commitRefs[commit.oid] = commitRefs
	}

	for _, branch := range commitRefs.branches {
		if branch.name == newBranch.name {
			return
		}
	}

	commitRefs.branches = append(commitRefs.branches, newBranch)
}

func (commitRefSet *commitRefSet) refsForCommit(commit *Commit) (commitRefsCopy *CommitRefs) {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefsCopy = &CommitRefs{}

	commitRefs, ok := commitRefSet.commitRefs[commit.oid]
	if ok {
		commitRefsCopy.tags = append([]*Tag(nil), commitRefs.tags...)
		commitRefsCopy.branches = append([]*Branch(nil), commitRefs.branches...)
	}

	return commitRefsCopy
}

type refCommitSets struct {
	commits  map[string]commitSet
	channels *Channels
	lock     sync.Mutex
}

func newRefCommitSets(channels *Channels) *refCommitSets {
	return &refCommitSets{
		commits:  make(map[string]commitSet),
		channels: channels,
	}
}

func (refCommitSets *refCommitSets) commitSet(ref Ref) (commitSet commitSet, exists bool) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, exists = refCommitSets.commits[ref.Name()]
	return
}

func (refCommitSets *refCommitSets) setCommitSet(ref Ref, commitSet commitSet) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	refCommitSets.commits[ref.Name()] = commitSet
}

func (refCommitSets *refCommitSets) addCommitFilter(ref Ref, commitFilter *CommitFilter) (err error) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, ok := refCommitSets.commits[ref.Name()]
	if !ok {
		return fmt.Errorf("No CommitSet exists for ref: %v", ref.Name())
	}

	filteredCommitSet := newFilteredCommitSet(commitSet, commitFilter)
	refCommitSets.commits[ref.Name()] = filteredCommitSet

	go func() {
		beforeState := commitSet.CommitSetState()
		filteredCommitSet.initialiseFromCommitSet()

		if !beforeState.loading {
			afterState := filteredCommitSet.CommitSetState()

			if afterState.commitNum < beforeState.commitNum {
				refCommitSets.channels.ReportStatus("Filter reduced %v commits to %v commits",
					beforeState.commitNum, afterState.commitNum)
			} else {
				refCommitSets.channels.ReportStatus("Filter had no effect")
			}
		}

	}()

	refCommitSets.channels.ReportStatus("Applying commit filter...")

	return
}

func (refCommitSets *refCommitSets) removeCommitFilter(ref Ref) (err error) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, ok := refCommitSets.commits[ref.Name()]
	if !ok {
		return fmt.Errorf("No CommitSet exists for ref: %v", ref.Name())
	}

	filteredCommitSet, ok := commitSet.(*filteredCommitSet)
	if !ok {
		log.Errorf("Unknown commitSet type %T", commitSet)
		return
	}

	if !filteredCommitSet.HasChild() {
		refCommitSets.channels.ReportStatus("No commit filter applied to remove")
		return
	}

	refCommitSets.commits[ref.Name()] = filteredCommitSet.Child()
	refCommitSets.channels.ReportStatus("Removed commit filter")

	return
}

type statusManager struct {
	repoDataLoader  *RepoDataLoader
	status          *Status
	statusListeners []StatusListener
	lock            sync.Mutex
}

func newStatusManager(repoDataLoader *RepoDataLoader) *statusManager {
	return &statusManager{
		repoDataLoader: repoDataLoader,
	}
}

func (statusManager *statusManager) loadStatus() (err error) {
	newStatus, err := statusManager.repoDataLoader.LoadStatus()
	if err != nil {
		return
	}

	statusManager.lock.Lock()
	defer statusManager.lock.Unlock()

	if statusManager.status == nil || !statusManager.status.Equal(newStatus) {
		log.Debugf("Git status has changed. Notifying status listeners.")
		statusManager.status = newStatus

		for _, statusListener := range statusManager.statusListeners {
			statusListener.OnStatusChanged(newStatus)
		}
	}

	return
}

func (statusManager *statusManager) getStatus() *Status {
	statusManager.lock.Lock()
	defer statusManager.lock.Unlock()

	return statusManager.status
}

func (statusManager *statusManager) registerStatusListener(statusListener StatusListener) {
	if statusListener == nil {
		return
	}

	log.Debugf("Registering status listener %T", statusListener)

	statusManager.lock.Lock()
	defer statusManager.lock.Unlock()

	statusManager.statusListeners = append(statusManager.statusListeners, statusListener)
}

// RepositoryData implements RepoData and stores all loaded repository data
type RepositoryData struct {
	channels       *Channels
	repoDataLoader *RepoDataLoader
	head           Ref
	refSet         *refSet
	commitRefSet   *commitRefSet
	refCommitSets  *refCommitSets
	statusManager  *statusManager
	refUpdateCh    chan *UpdatedRef
}

// NewRepositoryData creates a new instance
func NewRepositoryData(repoDataLoader *RepoDataLoader, channels *Channels) *RepositoryData {
	return &RepositoryData{
		channels:       channels,
		repoDataLoader: repoDataLoader,
		refSet:         newRefSet(),
		commitRefSet:   newCommitRefSet(),
		refCommitSets:  newRefCommitSets(channels),
		statusManager:  newStatusManager(repoDataLoader),
		refUpdateCh:    make(chan *UpdatedRef, updatedRefChannelSize),
	}
}

// Free free's any underlying resources
func (repoData *RepositoryData) Free() {
	close(repoData.refUpdateCh)
	repoData.repoDataLoader.Free()
}

// Initialise performs setup to allow loading data from the repository
func (repoData *RepositoryData) Initialise(repoPath string) (err error) {
	path, err := repoData.processPath(repoPath)
	if err != nil {
		return
	}

	if err = repoData.repoDataLoader.Initialise(path); err != nil {
		return
	}

	go repoData.processUpdatedRefs()
	repoData.RegisterRefStateListener(repoData)

	return repoData.LoadStatus()
}

func (repoData *RepositoryData) processPath(repoPath string) (processedPath string, err error) {
	path, err := CanonicalPath(repoPath)
	if err != nil {
		return
	}

	for {
		gitDirPath := filepath.Join(path, GitRepositoryDirectoryName)
		log.Debugf("gitDirPath: %v", gitDirPath)

		if _, err = os.Stat(gitDirPath); err != nil {
			if !os.IsNotExist(err) {
				break
			}
		} else {
			processedPath = gitDirPath
			break
		}

		if path == "/" {
			err = fmt.Errorf("Unable to find a git repository in %v or any of its parent directories", repoPath)
			break
		}

		path = filepath.Dir(path)
	}

	return
}

// Path returns the file patch location of the repository
func (repoData *RepositoryData) Path() string {
	return repoData.repoDataLoader.Path()
}

// LoadHead attempts to load the HEAD reference
func (repoData *RepositoryData) LoadHead() (err error) {
	head, err := repoData.repoDataLoader.Head()
	if err != nil {
		return
	}

	repoData.head = head

	return
}

// LoadRefs loads all branches and tags present in the repository
func (repoData *RepositoryData) LoadRefs(onRefsLoaded OnRefsLoaded) {
	refSet := repoData.refSet

	log.Debug("Loading refs")

	if !refSet.startUpdate() {
		log.Debugf("Already loading refs")
		return
	}

	go func() {
		defer refSet.endUpdate()

		if err := repoData.LoadHead(); err != nil {
			repoData.channels.ReportError(err)
			return
		}

		refs, err := repoData.repoDataLoader.LoadRefs()
		if err != nil {
			repoData.channels.ReportError(err)
			return
		}

		if err = repoData.mapRefsToCommits(refs); err != nil {
			repoData.channels.ReportError(err)
			return
		}

		if err = refSet.update(refs); err != nil {
			repoData.channels.ReportError(err)
			return
		}

		refSet.endUpdate()

		log.Debug("Refs loaded")

		if onRefsLoaded != nil {
			if err = onRefsLoaded(refs); err != nil {
				repoData.channels.ReportError(err)
			}
		}
	}()
}

// TODO Become RefStateListener and only update commitRefSet for refs that have changed
func (repoData *RepositoryData) mapRefsToCommits(refs []Ref) (err error) {
	log.Debug("Mapping refs to commits")

	var commit *Commit
	commitRefSet := repoData.commitRefSet

	commitRefSet.clear()

	for _, ref := range refs {
		commit, err = repoData.repoDataLoader.Commit(ref.Oid())
		if err != nil {
			return
		}

		switch refInstance := ref.(type) {
		case *Branch:
			commitRefSet.addBranchForCommit(commit, refInstance)
		case *Tag:
			commitRefSet.addTagForCommit(commit, refInstance)
		}
	}

	return
}

// LoadCommits attempts to load all commits for the provided oid
func (repoData *RepositoryData) LoadCommits(ref Ref, onCommitsLoaded OnCommitsLoaded) (err error) {
	if _, ok := repoData.refCommitSets.commitSet(ref); ok {
		log.Debugf("Commits already loading/loaded for ref %v", ref.Name())
		return
	}

	commitCh, err := repoData.repoDataLoader.Commits(ref.Oid())
	if err != nil {
		return
	}

	commitSet := newBaseFilteredCommitSet()
	commitSet.SetLoading(true)
	repoData.refCommitSets.setCommitSet(ref, commitSet)

	go func() {
		log.Debugf("Receiving commits from RepoDataLoader for ref %v at %v", ref.Name(), ref.Oid())

		for commit := range commitCh {
			commitSet, ok := repoData.refCommitSets.commitSet(ref)
			if !ok {
				log.Errorf("Error when loading commits: No CommitSet exists for ref %v", ref.Name())
				return
			}

			if err := commitSet.AddCommit(commit); err != nil {
				log.Errorf("Error when loading commits for ref %v: %v", ref.Name(), err)
				return
			}
		}

		commitSet, ok := repoData.refCommitSets.commitSet(ref)
		if !ok {
			log.Errorf("No CommitSet exists for ref %v", ref.Name())
			return
		}

		commitSet.SetLoading(false)
		log.Debugf("Finished loading commits for ref %v", ref.Name())

		repoData.channels.ReportError(onCommitsLoaded(ref))
	}()

	return
}

// Head returns the loaded HEAD ref
func (repoData *RepositoryData) Head() Ref {
	return repoData.head
}

// Branches returns all loaded local and remote branches
func (repoData *RepositoryData) Branches() (localBranches []*Branch, remoteBranches []*Branch, loading bool) {
	refSet := repoData.refSet
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	localBranches = refSet.localBranchesList
	remoteBranches = refSet.remoteBranchesList
	loading = refSet.loading

	return
}

// LocalTags returns all loaded tags
func (repoData *RepositoryData) LocalTags() (tags []*Tag, loading bool) {
	refSet := repoData.refSet
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	tags = refSet.tagsList
	loading = refSet.loading

	return
}

// RefsForCommit returns the set of all refs that point to the provided commit
func (repoData *RepositoryData) RefsForCommit(commit *Commit) *CommitRefs {
	return repoData.commitRefSet.refsForCommit(commit)
}

// CommitSetState returns the current commit set state for the provided oid
func (repoData *RepositoryData) CommitSetState(ref Ref) CommitSetState {
	if commitSet, ok := repoData.refCommitSets.commitSet(ref); ok {
		return commitSet.CommitSetState()
	}

	return CommitSetState{
		loading:   false,
		commitNum: 0,
	}
}

// Commits returns a channel from which the commit range specified can be read
func (repoData *RepositoryData) Commits(ref Ref, startIndex, count uint) (<-chan *Commit, error) {
	commitSet, ok := repoData.refCommitSets.commitSet(ref)
	if !ok {
		return nil, fmt.Errorf("No commits loaded for ref %v", ref.Name())
	}

	commitCh := make(chan *Commit)

	go func() {
		defer close(commitCh)
		var commit *Commit
		index := startIndex

		for {
			if index-startIndex < count {
				commit = commitSet.Commit(index)
			}

			if commit != nil {
				commitCh <- commit
				index++
				commit = nil
			} else {
				return
			}
		}
	}()

	return commitCh, nil
}

// CommitByIndex returns the loaded commit for the provided ref and index
func (repoData *RepositoryData) CommitByIndex(ref Ref, index uint) (commit *Commit, err error) {
	commitSet, ok := repoData.refCommitSets.commitSet(ref)
	if !ok {
		return nil, fmt.Errorf("No commits loaded for ref %v", ref.Name())
	}

	if commit = commitSet.Commit(index); commit == nil {
		err = fmt.Errorf("Commit index %v is invalid for ref %v", index, ref.Name())
	}

	return
}

// Commit loads the commit from the repository using the provided oid
func (repoData *RepositoryData) Commit(oid *Oid) (*Commit, error) {
	return repoData.repoDataLoader.Commit(oid)
}

// AddCommitFilter adds the filter to the specified ref
func (repoData *RepositoryData) AddCommitFilter(ref Ref, commitFilter *CommitFilter) error {
	return repoData.refCommitSets.addCommitFilter(ref, commitFilter)
}

// RemoveCommitFilter removes a filter (if one exists) for the specified oid
func (repoData *RepositoryData) RemoveCommitFilter(ref Ref) error {
	return repoData.refCommitSets.removeCommitFilter(ref)
}

// DiffCommit loads a diff between the commit with the specified oid and its parent
// If the commit has more than one parent no diff is returned
func (repoData *RepositoryData) DiffCommit(commit *Commit) (*Diff, error) {
	return repoData.repoDataLoader.DiffCommit(commit)
}

// DiffFile Generates a diff for the provided file
// If statusType is StStaged then the diff is between HEAD and the index
// If statusType is StUnstaged then the diff is between index and the working directory
func (repoData *RepositoryData) DiffFile(statusType StatusType, path string) (*Diff, error) {
	return repoData.repoDataLoader.DiffFile(statusType, path)
}

// LoadStatus loads the current git status
func (repoData *RepositoryData) LoadStatus() (err error) {
	log.Debugf("Loading git status")
	return repoData.statusManager.loadStatus()
}

// RegisterStatusListener registers a listener to be notified when git status changes
func (repoData *RepositoryData) RegisterStatusListener(statusListener StatusListener) {
	repoData.statusManager.registerStatusListener(statusListener)
}

// RegisterRefStateListener registers a listener to be notified when a ref is added, removed or modified
func (repoData *RepositoryData) RegisterRefStateListener(refStateListener RefStateListener) {
	repoData.refSet.registerRefStateListener(refStateListener)
}

// OnRefsChanged processes modified refs and loads any missing commits
func (repoData *RepositoryData) OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	for _, updatedRef := range updatedRefs {
		select {
		case repoData.refUpdateCh <- updatedRef:
		default:
			log.Errorf("Unable process UpdatedRef %v", updatedRef)
		}
	}
}

func (repoData *RepositoryData) processUpdatedRefs() {
	log.Info("Starting UpdatedRef processor")

	for updatedRef := range repoData.refUpdateCh {
		oldRef := updatedRef.OldRef
		newRef := updatedRef.NewRef

		log.Debugf("Processing ref update for %v", updatedRef)

		commitSet, exists := repoData.refCommitSets.commitSet(oldRef)
		if !exists {
			log.Debugf("No commitSet for oid %v", oldRef.Oid())
			continue
		}

		commonAncestor, err := repoData.repoDataLoader.MergeBase(newRef.Oid(), oldRef.Oid())
		if err != nil {
			log.Errorf("Unable to update commits for ref %v: %v", newRef.Name(), err)
			continue
		}
		log.Debugf("Common ancestor is %v", commonAncestor)

		commonAncestorCommit, err := repoData.repoDataLoader.Commit(commonAncestor)
		if err != nil {
			log.Errorf("Unable to load common ancestor commit with id %v: %v", commonAncestor, err)
			continue
		}

		commitRange := fmt.Sprintf("%v..%v", oldRef.Oid(), newRef.Oid())
		commitCh, err := repoData.repoDataLoader.CommitRange(commitRange)
		if err != nil {
			log.Errorf("Unable to load commits for range %v: %v", newRef.Name(), err)
			continue
		}
		log.Debugf("Reading commits for range %v", commitRange)

		var commits []*Commit
		for commit := range commitCh {
			commits = append(commits, commit)
		}

		if repoData.channels.Exit() {
			return
		}

		log.Debugf("Update ref %v with %v commits", newRef.Name(), len(commits))
		commitSet.Update(commonAncestorCommit, commits)

		repoData.refCommitSets.setCommitSet(newRef, commitSet)

		repoData.channels.UpdateDisplay()
	}
}
