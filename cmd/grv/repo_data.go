package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
)

const (
	// GitRepositoryDirectoryName is the name of the git directory in a git repository
	GitRepositoryDirectoryName = ".git"
)

// OnCommitsLoaded is called when all commits are loaded for the specified oid
type OnCommitsLoaded func(*Oid) error

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

// RefStateListener is updated when changes to refs are detected
type RefStateListener interface {
	OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef)
}

// RepoData houses all data loaded from the repository
type RepoData interface {
	Path() string
	LoadHead() error
	LoadRefs(OnRefsLoaded)
	LoadCommits(*Oid, OnCommitsLoaded) error
	Head() (*Oid, *Branch)
	Branches() (localBranches, remoteBranches []*Branch, loading bool)
	LocalTags() (tags []*Tag, loading bool)
	RefsForCommit(*Commit) *CommitRefs
	CommitSetState(*Oid) CommitSetState
	Commits(oid *Oid, startIndex, count uint) (<-chan *Commit, error)
	CommitByIndex(oid *Oid, index uint) (*Commit, error)
	Commit(oid *Oid) (*Commit, error)
	AddCommitFilter(*Oid, *CommitFilter) error
	RemoveCommitFilter(*Oid) error
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
}

type rawCommitSet struct {
	commits []*Commit
	loading bool
	lock    sync.Mutex
}

func newRawCommitSet() *rawCommitSet {
	return &rawCommitSet{
		commits: make([]*Commit, 0),
	}
}

// Add a commit to the commit set
func (rawCommitSet *rawCommitSet) AddCommit(commit *Commit) (err error) {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	if rawCommitSet.loading {
		rawCommitSet.commits = append(rawCommitSet.commits, commit)
	} else {
		err = fmt.Errorf("Cannot add commit when CommitSet is not in loading state")
	}

	return
}

// Commit returns the commit at the provided index (or nil if the index is invalid)
func (rawCommitSet *rawCommitSet) Commit(index uint) (commit *Commit) {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	if index < uint(len(rawCommitSet.commits)) {
		commit = rawCommitSet.commits[index]
	}

	return
}

// CommitStream returns a channel through which all the commits in this set can be read
func (rawCommitSet *rawCommitSet) CommitStream() <-chan *Commit {
	ch := make(chan *Commit)

	go func() {
		defer close(ch)
		var commit *Commit
		index := 0

		for {
			rawCommitSet.lock.Lock()

			length := len(rawCommitSet.commits)
			if index < length {
				commit = rawCommitSet.commits[index]
			}

			rawCommitSet.lock.Unlock()

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

// SetLoading sets whether this commit set is still loading or not
func (rawCommitSet *rawCommitSet) SetLoading(loading bool) {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	rawCommitSet.loading = loading
}

// CommitSetState returns the current state of the commit set
func (rawCommitSet *rawCommitSet) CommitSetState() CommitSetState {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	return CommitSetState{
		loading:   rawCommitSet.loading,
		commitNum: uint(len(rawCommitSet.commits)),
	}
}

type filteredCommitSet struct {
	commits      []*Commit
	commitSet    commitSet
	commitFilter *CommitFilter
	lock         sync.Mutex
}

func newFilteredCommitSet(commitSet commitSet, commitFilter *CommitFilter) *filteredCommitSet {
	return &filteredCommitSet{
		commits:      make([]*Commit, 0),
		commitSet:    commitSet,
		commitFilter: commitFilter,
	}
}

func (filteredCommitSet *filteredCommitSet) initialiseFromCommitSet() {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	for commit := range filteredCommitSet.commitSet.CommitStream() {
		filteredCommitSet.addCommitIfFilterMatches(commit)
	}
}

// CommitSet returns the child commit set of this filter
func (filteredCommitSet *filteredCommitSet) CommitSet() commitSet {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	return filteredCommitSet.commitSet
}

// AddCommit adds the commit to the child and then itself if the filter matches
func (filteredCommitSet *filteredCommitSet) AddCommit(commit *Commit) (err error) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if err = filteredCommitSet.commitSet.AddCommit(commit); err != nil {
		return
	}

	filteredCommitSet.addCommitIfFilterMatches(commit)

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
	filteredCommitSet.commitSet.SetLoading(loading)
}

// CommitSetState returns the state of this commit set
func (filteredCommitSet *filteredCommitSet) CommitSetState() CommitSetState {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	commitSetState := filteredCommitSet.commitSet.CommitSetState()

	if commitSetState.filterState == nil {
		commitSetState.filterState = &CommitSetFilterState{
			unfilteredCommitNum: commitSetState.commitNum,
		}
	}

	commitSetState.commitNum = uint(len(filteredCommitSet.commits))
	commitSetState.filterState.filtersApplied++

	return commitSetState
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
		log.Debugf("Refs Changed - new: %v, removed: %v, updated: %v")
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
	commits  map[*Oid]commitSet
	channels *Channels
	lock     sync.Mutex
}

func newRefCommitSets(channels *Channels) *refCommitSets {
	return &refCommitSets{
		commits:  make(map[*Oid]commitSet),
		channels: channels,
	}
}

func (refCommitSets *refCommitSets) commitSet(oid *Oid) (commitSet commitSet, exists bool) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, exists = refCommitSets.commits[oid]
	return
}

func (refCommitSets *refCommitSets) setCommitSet(oid *Oid, commitSet commitSet) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	refCommitSets.commits[oid] = commitSet
}

func (refCommitSets *refCommitSets) addCommitFilter(oid *Oid, commitFilter *CommitFilter) (err error) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, ok := refCommitSets.commits[oid]
	if !ok {
		return fmt.Errorf("No CommitSet exists for ref with id: %v", oid)
	}

	filteredCommitSet := newFilteredCommitSet(commitSet, commitFilter)
	refCommitSets.commits[oid] = filteredCommitSet

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

func (refCommitSets *refCommitSets) removeCommitFilter(oid *Oid) (err error) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, ok := refCommitSets.commits[oid]
	if !ok {
		return fmt.Errorf("No CommitSet exists for ref with id: %v", oid)
	}

	filteredCommitSet, ok := commitSet.(*filteredCommitSet)
	if !ok {
		refCommitSets.channels.ReportStatus("No commit filter applied to remove")
		return
	}

	refCommitSets.commits[oid] = filteredCommitSet.CommitSet()
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
	head           *Oid
	headBranch     *Branch
	refSet         *refSet
	commitRefSet   *commitRefSet
	refCommitSets  *refCommitSets
	statusManager  *statusManager
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
	}
}

// Free free's any underlying resources
func (repoData *RepositoryData) Free() {
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
	head, branch, err := repoData.repoDataLoader.Head()
	if err != nil {
		return
	}

	repoData.head = head
	repoData.headBranch = branch

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
func (repoData *RepositoryData) LoadCommits(oid *Oid, onCommitsLoaded OnCommitsLoaded) (err error) {
	if _, ok := repoData.refCommitSets.commitSet(oid); ok {
		log.Debugf("Commits already loading/loaded for oid %v", oid)
		return
	}

	commitCh, err := repoData.repoDataLoader.Commits(oid)
	if err != nil {
		return
	}

	commitSet := newRawCommitSet()
	commitSet.SetLoading(true)
	repoData.refCommitSets.setCommitSet(oid, commitSet)

	go func() {
		log.Debugf("Receiving commits from RepoDataLoader for oid %v", oid)

		for commit := range commitCh {
			commitSet, ok := repoData.refCommitSets.commitSet(oid)
			if !ok {
				log.Errorf("Error when loading commits: No CommitSet exists for ref with id: %v", oid)
				return
			}

			if err := commitSet.AddCommit(commit); err != nil {
				log.Errorf("Error when loading commits for oid %v: %v", oid, err)
				return
			}
		}

		commitSet, ok := repoData.refCommitSets.commitSet(oid)
		if !ok {
			log.Errorf("No CommitSet exists for ref with id: %v", oid)
			return
		}

		commitSet.SetLoading(false)
		log.Debugf("Finished loading commits for oid %v", oid)

		repoData.channels.ReportError(onCommitsLoaded(oid))
	}()

	return
}

// Head returns the loaded HEAD ref
func (repoData *RepositoryData) Head() (*Oid, *Branch) {
	return repoData.head, repoData.headBranch
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
func (repoData *RepositoryData) CommitSetState(oid *Oid) CommitSetState {
	if commitSet, ok := repoData.refCommitSets.commitSet(oid); ok {
		return commitSet.CommitSetState()
	}

	return CommitSetState{
		loading:   false,
		commitNum: 0,
	}
}

// Commits returns a channel from which the commit range specified can be read
func (repoData *RepositoryData) Commits(oid *Oid, startIndex, count uint) (<-chan *Commit, error) {
	commitSet, ok := repoData.refCommitSets.commitSet(oid)
	if !ok {
		return nil, fmt.Errorf("No commits loaded for oid %v", oid)
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
func (repoData *RepositoryData) CommitByIndex(oid *Oid, index uint) (commit *Commit, err error) {
	commitSet, ok := repoData.refCommitSets.commitSet(oid)
	if !ok {
		return nil, fmt.Errorf("No commits loaded for oid %v", oid)
	}

	if commit = commitSet.Commit(index); commit == nil {
		err = fmt.Errorf("Commit index %v is invalid for branch %v", index, oid)
	}

	return
}

// Commit loads the commit from the repository using the provided oid
func (repoData *RepositoryData) Commit(oid *Oid) (*Commit, error) {
	return repoData.repoDataLoader.Commit(oid)
}

// AddCommitFilter adds the filter to the specified ref
func (repoData *RepositoryData) AddCommitFilter(oid *Oid, commitFilter *CommitFilter) error {
	return repoData.refCommitSets.addCommitFilter(oid, commitFilter)
}

// RemoveCommitFilter removes a filter (if one exists) for the specified oid
func (repoData *RepositoryData) RemoveCommitFilter(oid *Oid) error {
	return repoData.refCommitSets.removeCommitFilter(oid)
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
