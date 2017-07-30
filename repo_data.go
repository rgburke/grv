package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/bradfitz/slice"
	"sync"
)

type OnCommitsLoaded func(*Oid) error
type OnBranchesLoaded func(localBranches, remoteBranches []*Branch) error
type OnTagsLoaded func([]*Tag) error

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

type CommitSet interface {
	AddCommit(commit *Commit) (err error)
	Commit(index uint) (commit *Commit)
	CommitStream() <-chan *Commit
	SetLoading(loading bool)
	CommitSetState() CommitSetState
}

type RawCommitSet struct {
	commits []*Commit
	loading bool
	lock    sync.Mutex
}

func NewRawCommitSet() *RawCommitSet {
	return &RawCommitSet{
		commits: make([]*Commit, 0),
	}
}

func (rawCommitSet *RawCommitSet) AddCommit(commit *Commit) (err error) {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	if rawCommitSet.loading {
		rawCommitSet.commits = append(rawCommitSet.commits, commit)
	} else {
		err = fmt.Errorf("Cannot add commit when CommitSet is not in loading state")
	}

	return
}

func (rawCommitSet *RawCommitSet) Commit(index uint) (commit *Commit) {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	if index < uint(len(rawCommitSet.commits)) {
		commit = rawCommitSet.commits[index]
	}

	return
}

func (rawCommitSet *RawCommitSet) CommitStream() <-chan *Commit {
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

func (rawCommitSet *RawCommitSet) SetLoading(loading bool) {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	rawCommitSet.loading = loading
}

func (rawCommitSet *RawCommitSet) CommitSetState() CommitSetState {
	rawCommitSet.lock.Lock()
	defer rawCommitSet.lock.Unlock()

	return CommitSetState{
		loading:   rawCommitSet.loading,
		commitNum: uint(len(rawCommitSet.commits)),
	}
}

type FilteredCommitSet struct {
	commits      []*Commit
	commitSet    CommitSet
	commitFilter *CommitFilter
	lock         sync.Mutex
}

func NewFilteredCommitSet(commitSet CommitSet, commitFilter *CommitFilter) *FilteredCommitSet {
	return &FilteredCommitSet{
		commits:      make([]*Commit, 0),
		commitSet:    commitSet,
		commitFilter: commitFilter,
	}
}

func (filteredCommitSet *FilteredCommitSet) InitialiseFromCommitSet() {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	for commit := range filteredCommitSet.commitSet.CommitStream() {
		filteredCommitSet.addCommitIfFilterMatches(commit)
	}
}

func (filteredCommitSet *FilteredCommitSet) CommitSet() CommitSet {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	return filteredCommitSet.commitSet
}

func (filteredCommitSet *FilteredCommitSet) AddCommit(commit *Commit) (err error) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if err = filteredCommitSet.commitSet.AddCommit(commit); err != nil {
		return
	}

	filteredCommitSet.addCommitIfFilterMatches(commit)

	return
}

func (filteredCommitSet *FilteredCommitSet) addCommitIfFilterMatches(commit *Commit) {
	if filteredCommitSet.commitFilter.MatchesFilter(commit) {
		filteredCommitSet.commits = append(filteredCommitSet.commits, commit)
	}

	return
}

func (filteredCommitSet *FilteredCommitSet) Commit(index uint) (commit *Commit) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if index < uint(len(filteredCommitSet.commits)) {
		commit = filteredCommitSet.commits[index]
	}

	return
}

func (filteredCommitSet *FilteredCommitSet) CommitStream() <-chan *Commit {
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

func (filteredCommitSet *FilteredCommitSet) SetLoading(loading bool) {
	filteredCommitSet.commitSet.SetLoading(loading)
}

func (filteredCommitSet *FilteredCommitSet) CommitSetState() CommitSetState {
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

type CommitSetState struct {
	loading     bool
	commitNum   uint
	filterState *CommitSetFilterState
}

type CommitSetFilterState struct {
	unfilteredCommitNum uint
	filtersApplied      uint
}

type BranchSet struct {
	branches           map[*Oid]*Branch
	localBranchesList  []*Branch
	remoteBranchesList []*Branch
	loading            bool
	lock               sync.Mutex
}

func NewBranchSet() *BranchSet {
	return &BranchSet{
		branches: make(map[*Oid]*Branch),
	}
}

type TagSet struct {
	tags     map[*Oid]*Tag
	tagsList []*Tag
	loading  bool
	lock     sync.Mutex
}

func NewTagSet() *TagSet {
	return &TagSet{
		tags: make(map[*Oid]*Tag),
	}
}

type CommitRefs struct {
	tags     []*Tag
	branches []*Branch
}

type CommitRefSet struct {
	commitRefs map[*Oid]*CommitRefs
	lock       sync.Mutex
}

func NewCommitRefSet() *CommitRefSet {
	return &CommitRefSet{
		commitRefs: make(map[*Oid]*CommitRefs),
	}
}

func (commitRefSet *CommitRefSet) AddTagForCommit(commit *Commit, newTag *Tag) {
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

func (commitRefSet *CommitRefSet) AddBranchForCommit(commit *Commit, newBranch *Branch) {
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

func (commitRefSet *CommitRefSet) RefsForCommit(commit *Commit) (commitRefsCopy *CommitRefs) {
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

type RefCommitSets struct {
	commits  map[*Oid]CommitSet
	channels *Channels
	lock     sync.Mutex
}

func NewRefCommitSets(channels *Channels) *RefCommitSets {
	return &RefCommitSets{
		commits:  make(map[*Oid]CommitSet),
		channels: channels,
	}
}

func (refCommitSets *RefCommitSets) CommitSet(oid *Oid) (commitSet CommitSet, exists bool) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, exists = refCommitSets.commits[oid]
	return
}

func (refCommitSets *RefCommitSets) SetCommitSet(oid *Oid, commitSet CommitSet) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	refCommitSets.commits[oid] = commitSet
}

func (refCommitSets *RefCommitSets) AddCommitFilter(oid *Oid, commitFilter *CommitFilter) (err error) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, ok := refCommitSets.commits[oid]
	if !ok {
		return fmt.Errorf("No CommitSet exists for ref with id: %v", oid)
	}

	filteredCommitSet := NewFilteredCommitSet(commitSet, commitFilter)
	refCommitSets.commits[oid] = filteredCommitSet

	go func() {
		beforeState := commitSet.CommitSetState()
		filteredCommitSet.InitialiseFromCommitSet()

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

func (refCommitSets *RefCommitSets) RemoveCommitFilter(oid *Oid) (err error) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSet, ok := refCommitSets.commits[oid]
	if !ok {
		return fmt.Errorf("No CommitSet exists for ref with id: %v", oid)
	}

	filteredCommitSet, ok := commitSet.(*FilteredCommitSet)
	if !ok {
		refCommitSets.channels.ReportStatus("No commit filter applied to remove")
		return
	}

	refCommitSets.commits[oid] = filteredCommitSet.CommitSet()
	refCommitSets.channels.ReportStatus("Removed commit filter")

	return
}

type RepositoryData struct {
	channels       *Channels
	repoDataLoader *RepoDataLoader
	head           *Oid
	headBranch     *Branch
	branches       *BranchSet
	localTags      *TagSet
	commitRefSet   *CommitRefSet
	refCommitSets  *RefCommitSets
}

func NewRepositoryData(repoDataLoader *RepoDataLoader, channels *Channels) *RepositoryData {
	return &RepositoryData{
		channels:       channels,
		repoDataLoader: repoDataLoader,
		branches:       NewBranchSet(),
		localTags:      NewTagSet(),
		commitRefSet:   NewCommitRefSet(),
		refCommitSets:  NewRefCommitSets(channels),
	}
}

func (repoData *RepositoryData) Free() {
	repoData.repoDataLoader.Free()
}

func (repoData *RepositoryData) Initialise(repoPath string) (err error) {
	if err = repoData.repoDataLoader.Initialise(repoPath); err != nil {
		return
	}

	return
}

func (repoData *RepositoryData) Path() string {
	return repoData.repoDataLoader.Path()
}

func (repoData *RepositoryData) LoadHead() (err error) {
	head, branch, err := repoData.repoDataLoader.Head()
	if err != nil {
		return
	}

	repoData.head = head
	repoData.headBranch = branch

	return
}

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

		commitRefSet.AddBranchForCommit(commit, branch)
	}

	return
}

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

		commitRefSet.AddTagForCommit(commit, tag)
	}

	return
}

func (repoData *RepositoryData) LoadCommits(oid *Oid, onCommitsLoaded OnCommitsLoaded) (err error) {
	if _, ok := repoData.refCommitSets.CommitSet(oid); ok {
		log.Debugf("Commits already loading/loaded for oid %v", oid)
		return
	}

	commitCh, err := repoData.repoDataLoader.Commits(oid)
	if err != nil {
		return
	}

	commitSet := NewRawCommitSet()
	commitSet.SetLoading(true)
	repoData.refCommitSets.SetCommitSet(oid, commitSet)

	go func() {
		log.Debugf("Receiving commits from RepoDataLoader for oid %v", oid)

		for commit := range commitCh {
			commitSet, ok := repoData.refCommitSets.CommitSet(oid)
			if !ok {
				log.Errorf("Error when loading commits: No CommitSet exists for ref with id: %v", oid)
				return
			}

			if err := commitSet.AddCommit(commit); err != nil {
				log.Errorf("Error when loading commits for oid %v: %v", oid, err)
				return
			}
		}

		commitSet, ok := repoData.refCommitSets.CommitSet(oid)
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

func (repoData *RepositoryData) Head() (*Oid, *Branch) {
	return repoData.head, repoData.headBranch
}

func (repoData *RepositoryData) Branches() (localBranches []*Branch, remoteBranches []*Branch, loading bool) {
	branchSet := repoData.branches
	branchSet.lock.Lock()
	defer branchSet.lock.Unlock()

	localBranches = branchSet.localBranchesList
	remoteBranches = branchSet.remoteBranchesList
	loading = branchSet.loading

	return
}

func (repoData *RepositoryData) LocalTags() (tags []*Tag, loading bool) {
	tagSet := repoData.localTags
	tagSet.lock.Lock()
	defer tagSet.lock.Unlock()

	tags = tagSet.tagsList
	loading = tagSet.loading

	return
}

func (repoData *RepositoryData) RefsForCommit(commit *Commit) *CommitRefs {
	return repoData.commitRefSet.RefsForCommit(commit)
}

func (repoData *RepositoryData) CommitSetState(oid *Oid) CommitSetState {
	if commitSet, ok := repoData.refCommitSets.CommitSet(oid); ok {
		return commitSet.CommitSetState()
	}

	return CommitSetState{
		loading:   false,
		commitNum: 0,
	}
}

func (repoData *RepositoryData) Commits(oid *Oid, startIndex, count uint) (<-chan *Commit, error) {
	commitSet, ok := repoData.refCommitSets.CommitSet(oid)
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

func (repoData *RepositoryData) CommitByIndex(oid *Oid, index uint) (commit *Commit, err error) {
	commitSet, ok := repoData.refCommitSets.CommitSet(oid)
	if !ok {
		return nil, fmt.Errorf("No commits loaded for oid %v", oid)
	}

	if commit = commitSet.Commit(index); commit == nil {
		err = fmt.Errorf("Commit index %v is invalid for branch %v", index, oid)
	}

	return
}

func (repoData *RepositoryData) Commit(oid *Oid) (*Commit, error) {
	return repoData.repoDataLoader.Commit(oid)
}

func (repoData *RepositoryData) AddCommitFilter(oid *Oid, commitFilter *CommitFilter) error {
	return repoData.refCommitSets.AddCommitFilter(oid, commitFilter)
}

func (repoData *RepositoryData) RemoveCommitFilter(oid *Oid) error {
	return repoData.refCommitSets.RemoveCommitFilter(oid)
}

func (repoData *RepositoryData) Diff(commit *Commit) (*Diff, error) {
	return repoData.repoDataLoader.Diff(commit)
}
