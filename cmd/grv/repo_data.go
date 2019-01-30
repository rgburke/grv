package main

import (
	"fmt"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
)

const (
	updatedRefChannelSize = 256
)

// OnRefsLoaded is called when all refs have been loaded and processed
type OnRefsLoaded func([]Ref) error

// ReloadResult is called when a reload of cached repository data has been completed
type ReloadResult func(err error)

// CommitSetListener is notified of load and update events for commit sets
type CommitSetListener interface {
	OnCommitsLoaded(Ref)
	OnCommitsUpdated(Ref)
}

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
	OnHeadChanged(oldHead, newHead Ref)
	OnTrackingBranchesUpdated(trackingBranches []*LocalBranch)
}

// RepoData houses all data loaded from the repository
type RepoData interface {
	EventListener
	Path() string
	RepositoryRootPath() string
	Workdir() string
	UserEditor() (string, error)
	GenerateGitCommandEnvironment() (env []string, rootDir string)
	Reload(ReloadResult)
	LoadHead() error
	LoadRefs(OnRefsLoaded)
	LoadCommits(Ref) error
	Head() Ref
	Ref(refName string) (Ref, error)
	Branches() (localBranches, remoteBranches []Branch, loading bool)
	Tags() (tags []*Tag, loading bool)
	LocalBranches(*RemoteBranch) []*LocalBranch
	RefsForCommit(*Commit) *CommitRefs
	CommitSetState(Ref) CommitSetState
	Commits(ref Ref, startIndex, count uint) (<-chan *Commit, error)
	CommitByIndex(ref Ref, index uint) (*Commit, error)
	Commit(oid *Oid) (*Commit, error)
	CommitByOid(oidStr string) (*Commit, error)
	CommitParents(oid *Oid) ([]*Commit, error)
	AddCommitFilter(Ref, *CommitFilter) error
	RemoveCommitFilter(Ref) error
	DiffCommit(commit *Commit) (*Diff, error)
	DiffFile(statusType StatusType, path string) (*Diff, error)
	DiffStage(statusType StatusType) (*Diff, error)
	LoadStatus() (err error)
	Status() *Status
	LoadRemotes() error
	Remotes() []string
	RegisterStatusListener(StatusListener)
	RegisterRefStateListener(RefStateListener)
	RegisterCommitSetListener(CommitSetListener)
}

type commitSet interface {
	AddCommit(commit *Commit) (err error)
	Commit(index uint) (commit *Commit)
	CommitStream() <-chan *Commit
	SetLoading(loading bool)
	CommitSetState() CommitSetState
	Update([]*Commit)
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

func (filteredCommitSet *filteredCommitSet) Update(commits []*Commit) {
	filteredCommitSet.lock.Lock()
	defer filteredCommitSet.lock.Unlock()

	if filteredCommitSet.hasChild() {
		filteredCommitSet.child.Update(commits)
		return
	}

	filteredCommitSet.commits = commits
	filteredCommitSet.loading = false
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

type trackingBranchState struct {
	localBranch  *LocalBranch
	remoteBranch *RemoteBranch
}

type trackingBranchUpdater interface {
	updateTrackingBranches(trackingBranchStates []*trackingBranchState) []*LocalBranch
}

type refSet struct {
	refs                          map[string]Ref
	refsShorthand                 map[string]Ref
	headRef                       Ref
	localBranchesList             []Branch
	remoteBranchesList            []Branch
	tagsList                      []*Tag
	remoteToLocalTrackingBranches map[string]map[string]bool
	loading                       bool
	refStateListeners             []RefStateListener
	trackingBranchUpdater         trackingBranchUpdater
	lock                          sync.Mutex
}

func newRefSet(trackingBranchUpdater trackingBranchUpdater) *refSet {
	return &refSet{
		refs:                          make(map[string]Ref),
		refsShorthand:                 make(map[string]Ref),
		remoteToLocalTrackingBranches: make(map[string]map[string]bool),
		trackingBranchUpdater:         trackingBranchUpdater,
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

func (refSet *refSet) unregisterRefStateListener(refStateListener RefStateListener) {
	if refStateListener == nil {
		return
	}

	log.Debugf("Unregistering ref state listener %T", refStateListener)

	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	for index, listener := range refSet.refStateListeners {
		if refStateListener == listener {
			refSet.refStateListeners = append(refSet.refStateListeners[:index], refSet.refStateListeners[index+1:]...)
			break
		}
	}
}

func (refSet *refSet) updateHead(head Ref) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	if branch, isLocalBranch := head.(*LocalBranch); isLocalBranch {
		if branchRef, ok := refSet.refs[branch.Name()]; ok {
			if branch.Oid().Equal(branchRef.Oid()) {
				head = branchRef
			}
		}
	}

	oldHead := refSet.headRef
	refSet.headRef = head

	if oldHead != nil && !oldHead.Equal(head) {
		refSet.notifyRefStateListenersHeadChanged(oldHead, head)
	}
}

func (refSet *refSet) head() Ref {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	return refSet.headRef
}

func (refSet *refSet) ref(refName string) (ref Ref, exists bool) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	if ref, exists = refSet.refs[refName]; exists {
		return
	}

	ref, exists = refSet.refsShorthand[refName]
	return
}

func (refSet *refSet) startRefUpdate() bool {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	if refSet.loading {
		return false
	}

	refSet.loading = true

	return true
}

func (refSet *refSet) endRefUpdate() {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	refSet.loading = false
}

func (refSet *refSet) updateRefs(refs []Ref) (err error) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	log.Debugf("Updating refs in refSet")

	if !refSet.loading {
		return fmt.Errorf("RefSet not in loading state")
	}

	var addedRefs, removedRefs []Ref
	var updatedRefs []*UpdatedRef

	refMap := make(map[string]Ref)
	refShorthandMap := make(map[string]Ref)
	remoteToLocalTrackingBranches := make(map[string]map[string]bool)
	var localBranches, remoteBranches []Branch
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
		refShorthandMap[ref.Shorthand()] = ref

		switch rawRef := ref.(type) {
		case *RemoteBranch:
			remoteBranches = append(remoteBranches, rawRef)
		case *LocalBranch:
			localBranches = append(localBranches, rawRef)

			if rawRef.IsTrackingBranch() {
				trackingBranches, ok := remoteToLocalTrackingBranches[rawRef.remoteBranch]
				if !ok {
					trackingBranches = make(map[string]bool)
					remoteToLocalTrackingBranches[rawRef.remoteBranch] = trackingBranches
				}

				trackingBranches[rawRef.Name()] = true

				if isExisting {
					if existingLocalBranch, isLocalBranch := existingRef.(*LocalBranch); isLocalBranch &&
						existingLocalBranch.IsTrackingBranch() {
						rawRef.UpdateAheadBehind(existingLocalBranch.ahead, existingLocalBranch.behind)
					}
				}
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

	oldRemoteToLocalTrackingBranches := refSet.remoteToLocalTrackingBranches

	refSet.refs = refMap
	refSet.refsShorthand = refShorthandMap
	refSet.localBranchesList = localBranches
	refSet.remoteBranchesList = remoteBranches
	refSet.tagsList = tags
	refSet.remoteToLocalTrackingBranches = remoteToLocalTrackingBranches

	trackingBranchStates := refSet.determineTrackingBranchesToUpdate(
		oldRemoteToLocalTrackingBranches, remoteToLocalTrackingBranches, updatedRefs)

	if len(addedRefs) > 0 || len(removedRefs) > 0 || len(updatedRefs) > 0 {
		refSet.notifyRefStateListenersRefsChanged(addedRefs, removedRefs, updatedRefs)
	} else {
		log.Debugf("No new, removed or modified refs")
	}

	trackingBranches := refSet.trackingBranchUpdater.updateTrackingBranches(trackingBranchStates)

	if len(trackingBranches) > 0 {
		refSet.notifyRefStateListenersTrackingBranchesUpdated(trackingBranches)
	}

	return
}

func (refSet *refSet) determineTrackingBranchesToUpdate(remoteToLocalOld, remoteToLocalNew map[string]map[string]bool,
	updatedRefs []*UpdatedRef) (trackingBranchStates []*trackingBranchState) {

	trackingBranchesToUpdate := make(map[string]bool)

	for remoteBranch, localBranches := range remoteToLocalNew {
		existingLocalBranches, isExistingRemoteBranch := remoteToLocalOld[remoteBranch]

		if isExistingRemoteBranch {
			for localBranch := range localBranches {
				if _, isExistingLocalbranch := existingLocalBranches[localBranch]; !isExistingLocalbranch {
					trackingBranchesToUpdate[localBranch] = true
				}
			}
		} else {
			for localBranch := range localBranches {
				trackingBranchesToUpdate[localBranch] = true
			}
		}
	}

	for _, updatedRef := range updatedRefs {
		switch branch := updatedRef.NewRef.(type) {
		case *LocalBranch:
			trackingBranchesToUpdate[branch.Name()] = true
		case *RemoteBranch:
			localBranches, ok := remoteToLocalNew[branch.Name()]
			if ok {
				for localBranch := range localBranches {
					trackingBranchesToUpdate[localBranch] = true
				}
			}
		}
	}

	for trackingBranch := range trackingBranchesToUpdate {
		branch, ok := refSet.refs[trackingBranch]
		if !ok {
			log.Errorf("Algorithm error: Expected ref %v to exist", trackingBranch)
			continue
		}

		localBranch, ok := branch.(*LocalBranch)
		if !ok {
			log.Errorf("Algorithm error: Expected ref %v to be a LocalBranch instance", trackingBranch)
			continue
		}

		branch, ok = refSet.refs[localBranch.remoteBranch]
		if !ok {
			log.Errorf("Algorithm error: Expected ref %v to exist", localBranch.remoteBranch)
			continue
		}

		remoteBranch, ok := branch.(*RemoteBranch)
		if !ok {
			log.Errorf("Algorithm error: Expected ref %v to be a RemoteBranch instance", localBranch.remoteBranch)
			continue
		}

		trackingBranchStates = append(trackingBranchStates, &trackingBranchState{
			localBranch:  localBranch,
			remoteBranch: remoteBranch,
		})
	}

	return
}

func (refSet *refSet) notifyRefStateListenersRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	refStateListeners := append([]RefStateListener(nil), refSet.refStateListeners...)

	go func() {
		log.Debugf("Notifying RefStateListeners Refs Changed - new: %v, removed: %v, updated: %v",
			len(addedRefs), len(removedRefs), len(updatedRefs))

		for _, refStateListener := range refStateListeners {
			refStateListener.OnRefsChanged(addedRefs, removedRefs, updatedRefs)
		}
	}()
}

func (refSet *refSet) notifyRefStateListenersHeadChanged(oldHead, newHead Ref) {
	refStateListeners := append([]RefStateListener(nil), refSet.refStateListeners...)

	go func() {
		log.Debugf("Notifying RefStateListeners HEAD changed %v:%v -> %v:%v",
			oldHead.Name(), oldHead.Oid(), newHead.Name(), newHead.Oid())

		for _, refStateListener := range refStateListeners {
			refStateListener.OnHeadChanged(oldHead, newHead)
		}
	}()
}

func (refSet *refSet) notifyRefStateListenersTrackingBranchesUpdated(trackingBranches []*LocalBranch) {
	refStateListeners := append([]RefStateListener(nil), refSet.refStateListeners...)

	go func() {
		log.Debugf("Notifying RefStateListeners %v tracking branches have changed", len(trackingBranches))

		for _, refStateListener := range refStateListeners {
			refStateListener.OnTrackingBranchesUpdated(trackingBranches)
		}
	}()
}

func (refSet *refSet) branches() (localBranchesList, remoteBranchesList []Branch, loading bool) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	localBranchesList = append(localBranchesList, refSet.localBranchesList...)
	remoteBranchesList = append(remoteBranchesList, refSet.remoteBranchesList...)
	loading = refSet.loading

	return
}

func (refSet *refSet) tags() (tagsList []*Tag, loading bool) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	tagsList = append(tagsList, refSet.tagsList...)
	loading = refSet.loading

	return
}

func (refSet *refSet) localBranches(remoteBranch *RemoteBranch) (localBranches []*LocalBranch) {
	refSet.lock.Lock()
	defer refSet.lock.Unlock()

	localBranchNames, exist := refSet.remoteToLocalTrackingBranches[remoteBranch.Name()]
	if !exist {
		return
	}

	for localBranchName := range localBranchNames {
		localBranch := refSet.refs[localBranchName].(*LocalBranch)
		localBranches = append(localBranches, localBranch)
	}

	return
}

// CommitRefs contain all refs to a commit
type CommitRefs struct {
	tags     []*Tag
	branches []Branch
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

func (commitRefSet *commitRefSet) addTagForCommit(oid *Oid, newTag *Tag) {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefs, ok := commitRefSet.commitRefs[oid]
	if !ok {
		commitRefs = &CommitRefs{}
		commitRefSet.commitRefs[oid] = commitRefs
	}

	for _, tag := range commitRefs.tags {
		if tag.name == newTag.name {
			return
		}
	}

	commitRefs.tags = append(commitRefs.tags, newTag)
}

func (commitRefSet *commitRefSet) addBranchForCommit(oid *Oid, newBranch Branch) {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefs, ok := commitRefSet.commitRefs[oid]
	if !ok {
		commitRefs = &CommitRefs{}
		commitRefSet.commitRefs[oid] = commitRefs
	}

	for _, branch := range commitRefs.branches {
		if branch.Name() == newBranch.Name() {
			return
		}
	}

	commitRefs.branches = append(commitRefs.branches, newBranch)
}

func (commitRefSet *commitRefSet) refsForCommit(oid *Oid) (commitRefsCopy *CommitRefs) {
	commitRefSet.lock.Lock()
	defer commitRefSet.lock.Unlock()

	commitRefsCopy = &CommitRefs{}

	commitRefs, ok := commitRefSet.commitRefs[oid]
	if ok {
		commitRefsCopy.tags = append([]*Tag(nil), commitRefs.tags...)
		commitRefsCopy.branches = append([]Branch(nil), commitRefs.branches...)
	}

	return commitRefsCopy
}

type refCommitSets struct {
	commits            map[string]commitSet
	commitSetListeners []CommitSetListener
	channels           Channels
	lock               sync.Mutex
}

func newRefCommitSets(channels Channels) *refCommitSets {
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

func (refCommitSets *refCommitSets) registerCommitSetListener(commitSetListener CommitSetListener) {
	if commitSetListener == nil {
		return
	}

	log.Debugf("Registering CommitSetListener %T", commitSetListener)

	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	refCommitSets.commitSetListeners = append(refCommitSets.commitSetListeners, commitSetListener)
}

func (refCommitSets *refCommitSets) unregisterCommitSetListener(commitSetListener CommitSetListener) {
	if commitSetListener == nil {
		return
	}

	log.Debugf("Unregistering CommitSetListener %T", commitSetListener)

	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	for index, listener := range refCommitSets.commitSetListeners {
		if commitSetListener == listener {
			refCommitSets.commitSetListeners = append(refCommitSets.commitSetListeners[:index], refCommitSets.commitSetListeners[index+1:]...)
			break
		}
	}
}

func (refCommitSets *refCommitSets) notifyCommitSetListenersCommitSetLoaded(ref Ref) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSetListeners := append([]CommitSetListener(nil), refCommitSets.commitSetListeners...)

	go func() {
		log.Debugf("Notifying CommitSetListeners commits for ref %v have loaded", ref.Name())

		for _, listener := range commitSetListeners {
			listener.OnCommitsLoaded(ref)
		}
	}()
}

func (refCommitSets *refCommitSets) notifyCommitSetListenersCommitSetUpdated(ref Ref) {
	refCommitSets.lock.Lock()
	defer refCommitSets.lock.Unlock()

	commitSetListeners := append([]CommitSetListener(nil), refCommitSets.commitSetListeners...)

	go func() {
		log.Debugf("Notifying CommitSetListeners commits for ref %v have updated", ref.Name())

		for _, listener := range commitSetListeners {
			listener.OnCommitsUpdated(ref)
		}
	}()
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

		statusListeners := append([]StatusListener(nil), statusManager.statusListeners...)
		go func() {
			for _, statusListener := range statusListeners {
				statusListener.OnStatusChanged(newStatus)
			}
		}()
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

func (statusManager *statusManager) unregisterStatusListener(statusListener StatusListener) {
	if statusListener == nil {
		return
	}

	log.Debugf("Unregistering status listener %T", statusListener)

	statusManager.lock.Lock()
	defer statusManager.lock.Unlock()

	for index, listener := range statusManager.statusListeners {
		if statusListener == listener {
			statusManager.statusListeners = append(statusManager.statusListeners[:index], statusManager.statusListeners[index+1:]...)
			break
		}
	}
}

type remoteSet struct {
	remotes []string
	lock    sync.Mutex
}

func newRemoteSet() *remoteSet {
	return &remoteSet{}
}

func (remoteSet *remoteSet) setRemotes(remotes []string) {
	remoteSet.lock.Lock()
	defer remoteSet.lock.Unlock()

	remoteSet.remotes = remotes
}

func (remoteSet *remoteSet) getRemotes() []string {
	remoteSet.lock.Lock()
	defer remoteSet.lock.Unlock()

	return remoteSet.remotes
}

// RepositoryData implements RepoData and stores all loaded repository data
type RepositoryData struct {
	channels       Channels
	repoDataLoader *RepoDataLoader
	head           Ref
	refSet         *refSet
	commitRefSet   *commitRefSet
	refCommitSets  *refCommitSets
	statusManager  *statusManager
	refUpdateCh    chan *UpdatedRef
	variables      *GRVVariables
	remoteSet      *remoteSet
	waitGroup      sync.WaitGroup
}

// NewRepositoryData creates a new instance
func NewRepositoryData(repoDataLoader *RepoDataLoader, channels Channels, variables *GRVVariables) *RepositoryData {
	repoData := &RepositoryData{
		channels:       channels,
		repoDataLoader: repoDataLoader,
		commitRefSet:   newCommitRefSet(),
		refCommitSets:  newRefCommitSets(channels),
		statusManager:  newStatusManager(repoDataLoader),
		refUpdateCh:    make(chan *UpdatedRef, updatedRefChannelSize),
		variables:      variables,
		remoteSet:      newRemoteSet(),
	}

	repoData.refSet = newRefSet(repoData)

	return repoData
}

// Free free's any underlying resources
func (repoData *RepositoryData) Free() {
	close(repoData.refUpdateCh)
	repoData.refUpdateCh = nil
	repoData.waitGroup.Wait()
}

// Initialise performs setup to allow loading data from the repository
func (repoData *RepositoryData) Initialise(repoSupplier RepoSupplier) (err error) {
	repoData.repoDataLoader.Initialise(repoSupplier)

	repoData.variables.SetVariable(VarRepoPath, repoData.Path())
	repoData.variables.SetVariable(VarRepoWorkDir, repoData.Workdir())

	repoData.waitGroup.Add(1)
	go repoData.processUpdatedRefs()
	repoData.RegisterRefStateListener(repoData)

	return repoData.LoadHead()
}

// Path returns the file path location of the repository
func (repoData *RepositoryData) Path() string {
	return repoData.repoDataLoader.Path()
}

// RepositoryRootPath returns the root working directory of the repository
func (repoData *RepositoryData) RepositoryRootPath() string {
	_, rootDir := repoData.GenerateGitCommandEnvironment()
	return rootDir
}

// Workdir returns working directory file path for the repository
func (repoData *RepositoryData) Workdir() string {
	return repoData.repoDataLoader.Workdir()
}

// UserEditor returns the editor git is configured to use
func (repoData *RepositoryData) UserEditor() (string, error) {
	return repoData.repoDataLoader.UserEditor()
}

// Reload cached repository data
func (repoData *RepositoryData) Reload(reloadResult ReloadResult) {
	notifyResult := func(err error) {
		if reloadResult != nil {
			reloadResult(err)
		}
	}

	go func() {
		if repoData.refSet.startRefUpdate() {
			err := repoData.loadRefs(nil)
			repoData.refSet.endRefUpdate()

			if err != nil {
				notifyResult(err)
				return
			}
		}

		if err := repoData.LoadStatus(); err != nil {
			notifyResult(err)
			return
		}

		if err := repoData.LoadRemotes(); err != nil {
			notifyResult(err)
			return
		}

		notifyResult(nil)
	}()
}

// LoadHead attempts to load the HEAD reference
func (repoData *RepositoryData) LoadHead() (err error) {
	head, err := repoData.repoDataLoader.Head()
	if err != nil {
		return
	}

	repoData.refSet.updateHead(head)

	if _, isDetached := head.(*HEAD); isDetached {
		repoData.variables.SetVariable(VarHead, head.Oid().String())
	} else {
		repoData.variables.SetVariable(VarHead, head.Name())
	}

	return
}

// LoadRefs loads all branches and tags present in the repository
func (repoData *RepositoryData) LoadRefs(onRefsLoaded OnRefsLoaded) {
	refSet := repoData.refSet

	log.Debug("Loading refs")

	if !refSet.startRefUpdate() {
		log.Debugf("Already loading refs")
		return
	}

	go func() {
		defer refSet.endRefUpdate()
		if err := repoData.loadRefs(onRefsLoaded); err != nil {
			repoData.channels.ReportError(err)
		}
	}()
}

func (repoData *RepositoryData) loadRefs(onRefsLoaded OnRefsLoaded) (err error) {
	refs, err := repoData.repoDataLoader.LoadRefs()
	if err != nil {
		return
	}

	if err = repoData.refSet.updateRefs(refs); err != nil {
		return
	}

	if err = repoData.LoadHead(); err != nil {
		return
	}

	log.Debug("Refs loaded")

	if onRefsLoaded != nil {
		err = onRefsLoaded(refs)
	}

	repoData.mapRefsToCommits(refs)
	repoData.channels.UpdateDisplay()

	return
}

// TODO Become RefStateListener and only update commitRefSet for refs that have changed
func (repoData *RepositoryData) mapRefsToCommits(refs []Ref) {
	log.Debug("Mapping refs to commits")

	commitRefSet := repoData.commitRefSet
	commitRefSet.clear()

	for _, ref := range refs {
		switch refInstance := ref.(type) {
		case Branch:
			commitRefSet.addBranchForCommit(ref.Oid(), refInstance)
		case *Tag:
			if commit, err := repoData.repoDataLoader.Commit(ref.Oid()); err != nil {
				log.Errorf("Unable to load commit for tag: %v", ref.Name())
			} else {
				commitRefSet.addTagForCommit(commit.oid, refInstance)
			}
		}
	}

	return
}

// LoadCommits attempts to load all commits for the provided oid
func (repoData *RepositoryData) LoadCommits(ref Ref) (err error) {
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

		repoData.refCommitSets.notifyCommitSetListenersCommitSetLoaded(ref)
	}()

	return
}

// Head returns the loaded HEAD ref
func (repoData *RepositoryData) Head() Ref {
	return repoData.refSet.head()
}

// Ref returns a ref instance (if one exists) identified by the provided name
func (repoData *RepositoryData) Ref(refName string) (ref Ref, err error) {
	ref, exists := repoData.refSet.ref(refName)
	if !exists {
		err = fmt.Errorf("No ref exists with name %v", refName)
	}

	return
}

// Branches returns all loaded local and remote branches
func (repoData *RepositoryData) Branches() (localBranches []Branch, remoteBranches []Branch, loading bool) {
	return repoData.refSet.branches()
}

// Tags returns all loaded tags
func (repoData *RepositoryData) Tags() (tags []*Tag, loading bool) {
	return repoData.refSet.tags()
}

// LocalBranches returns all local tracking branches for the provided remote branch
func (repoData *RepositoryData) LocalBranches(remoteBranch *RemoteBranch) []*LocalBranch {
	return repoData.refSet.localBranches(remoteBranch)
}

// RefsForCommit returns the set of all refs that point to the provided commit
func (repoData *RepositoryData) RefsForCommit(commit *Commit) *CommitRefs {
	return repoData.commitRefSet.refsForCommit(commit.oid)
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

// CommitByOid loads the commit from the repository using the provided oid string
func (repoData *RepositoryData) CommitByOid(oidStr string) (*Commit, error) {
	return repoData.repoDataLoader.CommitByOid(oidStr)
}

// CommitParents loads the parents of a commit
func (repoData *RepositoryData) CommitParents(oid *Oid) (parentCommits []*Commit, err error) {
	commit, err := repoData.repoDataLoader.Commit(oid)
	if err != nil {
		return
	}

	parentCount := commit.commit.ParentCount()
	var parentCommit *Commit

	for i := uint(0); i < parentCount; i++ {
		parentOid := &Oid{commit.commit.ParentId(i)}
		parentCommit, err = repoData.Commit(parentOid)
		if err != nil {
			return
		}

		parentCommits = append(parentCommits, parentCommit)
	}

	return
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

// DiffStage returns a diff for all files in the provided stage
func (repoData *RepositoryData) DiffStage(statusType StatusType) (*Diff, error) {
	return repoData.repoDataLoader.DiffStage(statusType)
}

// LoadStatus loads the current git status
func (repoData *RepositoryData) LoadStatus() (err error) {
	return repoData.statusManager.loadStatus()
}

// Status returns the current git status
func (repoData *RepositoryData) Status() *Status {
	return repoData.statusManager.getStatus()
}

// LoadRemotes loads remotes for the repository
func (repoData *RepositoryData) LoadRemotes() (err error) {
	remotes, err := repoData.repoDataLoader.Remotes()
	if err != nil {
		return
	}

	repoData.remoteSet.setRemotes(remotes)
	return
}

// Remotes returns remotes for the repository
func (repoData *RepositoryData) Remotes() []string {
	return repoData.remoteSet.getRemotes()
}

// RegisterStatusListener registers a listener to be notified when git status changes
func (repoData *RepositoryData) RegisterStatusListener(statusListener StatusListener) {
	repoData.statusManager.registerStatusListener(statusListener)
}

// RegisterRefStateListener registers a listener to be notified when a ref is added, removed or modified
func (repoData *RepositoryData) RegisterRefStateListener(refStateListener RefStateListener) {
	repoData.refSet.registerRefStateListener(refStateListener)
}

// RegisterCommitSetListener registers a listener to be notified when a commitSet event occurs
func (repoData *RepositoryData) RegisterCommitSetListener(commitSetListener CommitSetListener) {
	repoData.refCommitSets.registerCommitSetListener(commitSetListener)
}

// OnHeadChanged does nothing
func (repoData *RepositoryData) OnHeadChanged(oldHead, newHead Ref) {

}

// OnTrackingBranchesUpdated does nothing
func (repoData *RepositoryData) OnTrackingBranchesUpdated(trackingBranches []*LocalBranch) {

}

// OnRefsChanged processes modified refs and loads any missing commits
func (repoData *RepositoryData) OnRefsChanged(addedRefs, removedRefs []Ref, updatedRefs []*UpdatedRef) {
	repoData.addUpdatedRefsToProcessingQueue(updatedRefs)
}

func (repoData *RepositoryData) addUpdatedRefsToProcessingQueue(updatedRefs []*UpdatedRef) {
	refUpdateCh := repoData.refUpdateCh
	if refUpdateCh == nil {
		return
	}

	for _, updatedRef := range updatedRefs {
		select {
		case refUpdateCh <- updatedRef:
		default:
			log.Errorf("Unable process UpdatedRef %v", updatedRef)
		}
	}
}

func (repoData *RepositoryData) processUpdatedRefs() {
	defer repoData.waitGroup.Done()
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

		commitCh, err := repoData.repoDataLoader.Commits(newRef.Oid())
		if err != nil {
			log.Errorf("Unable to load commits for range %v: %v", newRef.Name(), err)
			continue
		}
		log.Debugf("Reading commits for oid %v", newRef.Oid())

		var commits []*Commit
		for commit := range commitCh {
			commits = append(commits, commit)

			if repoData.channels.Exit() {
				return
			}
		}

		log.Debugf("Updating ref %v with %v commits", newRef.Name(), len(commits))
		commitSet.Update(commits)
		repoData.refCommitSets.setCommitSet(newRef, commitSet)
		repoData.refCommitSets.notifyCommitSetListenersCommitSetUpdated(newRef)
		repoData.channels.UpdateDisplay()
	}
}

func (repoData *RepositoryData) updateTrackingBranches(trackingBranchStates []*trackingBranchState) (trackingBranches []*LocalBranch) {
	for _, trackingBranchState := range trackingBranchStates {
		localBranch := trackingBranchState.localBranch
		remoteBranch := trackingBranchState.remoteBranch

		ahead, behind, err := repoData.repoDataLoader.AheadBehind(localBranch.Oid(), remoteBranch.Oid())

		if err != nil {
			log.Errorf("Unable to determine ahead-behind counts for ref %v: %v", localBranch.Name(), err)
			continue
		}

		trackingBranchState.localBranch.UpdateAheadBehind(uint(ahead), uint(behind))
		trackingBranches = append(trackingBranches, localBranch)

		log.Debugf("%v is %v commits ahead and %v commits behind %v",
			localBranch.Name(), ahead, behind, remoteBranch.Name())
	}

	return
}

// HandleEvent reacts to an event
func (repoData *RepositoryData) HandleEvent(event Event) (err error) {
	switch event.EventType {
	case ViewRemovedEvent:
		repoData.handleViewRemovedEvent(event)
	}

	return
}

func (repoData *RepositoryData) handleViewRemovedEvent(event Event) {
	for _, view := range event.Args {
		if statusListener, ok := view.(StatusListener); ok {
			repoData.statusManager.unregisterStatusListener(statusListener)
		}

		if refStateListener, ok := view.(RefStateListener); ok {
			repoData.refSet.unregisterRefStateListener(refStateListener)
		}

		if commitSetListener, ok := view.(CommitSetListener); ok {
			repoData.refCommitSets.unregisterCommitSetListener(commitSetListener)
		}
	}
}

// GenerateGitCommandEnvironment populates git environment variables for
// the current repository
func (repoData *RepositoryData) GenerateGitCommandEnvironment() (env []string, rootDir string) {
	return repoData.repoDataLoader.GenerateGitCommandEnvironment()
}
