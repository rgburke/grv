package main

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
)

// OnCommitsLoaded is called when all commits are loaded for the specified oid
type OnCommitsLoaded func(*Oid) error

// OnBranchesLoaded is called when all local and remote branch refs have been loaded
type OnBranchesLoaded func(localBranches, remoteBranches []*Branch) error

// OnTagsLoaded is called when all tags have been loaded
type OnTagsLoaded func([]*Tag) error

// RepoData houses all data loaded from the repository
type RepoData interface {
	Path() string
	LoadHead() error
	LoadBranches(OnBranchesLoaded) error
	LoadLocalTags(OnTagsLoaded) error
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
	Diff(commit *Commit) (*Diff, error)
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

type branchSet struct {
	branches           map[*Oid]*Branch
	localBranchesList  []*Branch
	remoteBranchesList []*Branch
	loading            bool
	lock               sync.Mutex
}

func newBranchSet() *branchSet {
	return &branchSet{
		branches: make(map[*Oid]*Branch),
	}
}

type tagSet struct {
	tags     map[*Oid]*Tag
	tagsList []*Tag
	loading  bool
	lock     sync.Mutex
}

func newTagSet() *tagSet {
	return &tagSet{
		tags: make(map[*Oid]*Tag),
	}
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
	return &commitRefSet{
		commitRefs: make(map[*Oid]*CommitRefs),
	}
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

// RepositoryData implements RepoData and stores all loaded repository data
type RepositoryData struct {
	channels       *Channels
	repoDataLoader *RepoDataLoader
	head           *Oid
	headBranch     *Branch
	branches       *branchSet
	localTags      *tagSet
	commitRefSet   *commitRefSet
	refCommitSets  *refCommitSets
}

// NewRepositoryData creates a new instance
func NewRepositoryData(repoDataLoader *RepoDataLoader, channels *Channels) *RepositoryData {
	return &RepositoryData{
		channels:       channels,
		repoDataLoader: repoDataLoader,
		branches:       newBranchSet(),
		localTags:      newTagSet(),
		commitRefSet:   newCommitRefSet(),
		refCommitSets:  newRefCommitSets(channels),
	}
}

// Free free's any underlying resources
func (repoData *RepositoryData) Free() {
	repoData.repoDataLoader.Free()
}

// Initialise performs setup to allow loading data from the repository
func (repoData *RepositoryData) Initialise(repoPath string) (err error) {
	if err = repoData.repoDataLoader.Initialise(repoPath); err != nil {
		return
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

// LoadBranches attempts to load local and remote branch refs from the repository
func (repoData *RepositoryData) LoadBranches(onBranchesLoaded OnBranchesLoaded) (err error) {
	branchSet := repoData.branches
	branchSet.lock.Lock()
	defer branchSet.lock.Unlock()

	if branchSet.loading {
		log.Debug("Local branches already loading")
		return
	}

	go func() {
		branches, err := repoData.repoDataLoader.LoadBranches()
		if err != nil {
			repoData.channels.ReportError(err)
			return
		}

		slice.Sort(branches, func(i, j int) bool {
			return branches[i].name < branches[j].name
		})

		var localBranchesList []*Branch
		var remoteBranchesList []*Branch

		for _, branch := range branches {
			if branch.isRemote {
				remoteBranchesList = append(remoteBranchesList, branch)
			} else {
				localBranchesList = append(localBranchesList, branch)
			}
		}

		branchMap := make(map[*Oid]*Branch)
		for _, branch := range branches {
			branchMap[branch.oid] = branch
		}

		branchSet.lock.Lock()
		branchSet.branches = branchMap
		branchSet.localBranchesList = localBranchesList
		branchSet.remoteBranchesList = remoteBranchesList
		branchSet.loading = false
		branchSet.lock.Unlock()

		repoData.channels.ReportError(repoData.mapBranchesToCommits())
		repoData.channels.ReportError(onBranchesLoaded(localBranchesList, remoteBranchesList))
	}()

	branchSet.loading = true

	return
}

func (repoData *RepositoryData) mapBranchesToCommits() (err error) {
	branchSet := repoData.branches
	branchSet.lock.Lock()
	defer branchSet.lock.Unlock()

	commitRefSet := repoData.commitRefSet

	branches := append(branchSet.localBranchesList, branchSet.remoteBranchesList...)

	for _, branch := range branches {
		var commit *Commit
		commit, err = repoData.repoDataLoader.Commit(branch.oid)
		if err != nil {
			return
		}

		commitRefSet.addBranchForCommit(commit, branch)
	}

	return
}

// LoadLocalTags attempts to load all tags stored in the repository
func (repoData *RepositoryData) LoadLocalTags(onTagsLoaded OnTagsLoaded) (err error) {
	tagSet := repoData.localTags
	tagSet.lock.Lock()
	defer tagSet.lock.Unlock()

	if tagSet.loading {
		log.Debug("Local tags already loading")
		return
	}

	go func() {
		tags, err := repoData.repoDataLoader.LocalTags()
		if err != nil {
			repoData.channels.ReportError(err)
			return
		}

		slice.Sort(tags, func(i, j int) bool {
			return tags[i].name < tags[j].name
		})

		tagMap := make(map[*Oid]*Tag)
		for _, tag := range tags {
			tagMap[tag.oid] = tag
		}

		tagSet.lock.Lock()
		tagSet.tags = tagMap
		tagSet.tagsList = tags
		tagSet.loading = false
		tagSet.lock.Unlock()

		repoData.channels.ReportError(repoData.mapTagsToCommits())
		repoData.channels.ReportError(onTagsLoaded(tags))
	}()

	tagSet.loading = true

	return
}

func (repoData *RepositoryData) mapTagsToCommits() (err error) {
	tagSet := repoData.localTags
	tagSet.lock.Lock()
	defer tagSet.lock.Unlock()

	commitRefSet := repoData.commitRefSet

	for _, tag := range tagSet.tagsList {
		var commit *Commit
		commit, err = repoData.repoDataLoader.Commit(tag.oid)
		if err != nil {
			return
		}

		commitRefSet.addTagForCommit(commit, tag)
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
	branchSet := repoData.branches
	branchSet.lock.Lock()
	defer branchSet.lock.Unlock()

	localBranches = branchSet.localBranchesList
	remoteBranches = branchSet.remoteBranchesList
	loading = branchSet.loading

	return
}

// LocalTags returns all loaded tags
func (repoData *RepositoryData) LocalTags() (tags []*Tag, loading bool) {
	tagSet := repoData.localTags
	tagSet.lock.Lock()
	defer tagSet.lock.Unlock()

	tags = tagSet.tagsList
	loading = tagSet.loading

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

// Diff loads a diff for the specified oid
func (repoData *RepositoryData) Diff(commit *Commit) (*Diff, error) {
	return repoData.repoDataLoader.Diff(commit)
}
