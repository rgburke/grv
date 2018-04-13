package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
	git "gopkg.in/libgit2/git2go.v26"
)

const (
	// RdlHeadRef is the HEAD ref name
	RdlHeadRef          = "HEAD"
	rdlCommitBufferSize = 100
	rdlDiffStatsCols    = 80
	rdlShortOidLen      = 7
)

type instanceCache struct {
	oids       map[string]*Oid
	commits    map[string]*Commit
	oidLock    sync.Mutex
	commitLock sync.Mutex
}

// RepoDataLoader handles loading data from the repository
type RepoDataLoader struct {
	repo     *git.Repository
	cache    *instanceCache
	channels *Channels
}

// Oid is reference to a git object
type Oid struct {
	oid *git.Oid
}

// Equal returns true if this oid is equal to the provided oid
func (oid *Oid) Equal(other *Oid) bool {
	if other == nil {
		return false
	}

	return oid.oid.Cmp(other.oid) == 0
}

// String returns the oid hash
func (oid Oid) String() string {
	return oid.oid.String()
}

// ShortID returns a shortened oid hash
func (oid Oid) ShortID() (shortID string) {
	id := oid.String()

	if len(id) >= rdlShortOidLen {
		shortID = id[0:rdlShortOidLen]
	}

	return
}

// Ref is a named pointer to a commit
type Ref interface {
	Oid() *Oid
	Name() string
	Shorthand() string
	Equal(other Ref) bool
}

// Branch represents a branch reference
type Branch interface {
	Ref
	IsRemote() bool
}

type abstractBranch struct {
	oid       *Oid
	name      string
	shorthand string
}

// Oid pointed to by this branch
func (branch *abstractBranch) Oid() *Oid {
	return branch.oid
}

// Name of this branch
func (branch *abstractBranch) Name() string {
	return branch.name
}

// Shorthand name of this branch
func (branch *abstractBranch) Shorthand() string {
	return branch.shorthand
}

// Equal returns true if the other ref is a branch equal to this one
func (branch *abstractBranch) Equal(other Ref) bool {
	if other == nil {
		return false
	}

	otherBranch, ok := other.(*abstractBranch)
	if !ok {
		return false
	}

	return branch.Name() == otherBranch.Name() &&
		branch.Oid().Equal(otherBranch.Oid())
}

// String returns branch data in a string format
func (branch *abstractBranch) String() string {
	return fmt.Sprintf("%v:%v", branch.name, branch.oid)
}

// LocalBranch contains data for a local branch reference
type LocalBranch struct {
	*abstractBranch
	remoteBranch string
	ahead        uint
	behind       uint
}

func newLocalBranch(oid *Oid, rawBranch *git.Branch) (localBranch *LocalBranch, err error) {
	shorthand, err := rawBranch.Name()
	if err != nil {
		return
	}

	var upstreamRef string
	upstream, err := rawBranch.Upstream()
	if err != nil {
		if gitError, isGitError := err.(*git.GitError); !isGitError || gitError.Code != git.ErrNotFound {
			return
		}

		err = nil
	} else {
		upstreamRef = upstream.Name()
	}

	name := rawBranch.Reference.Name()

	localBranch = &LocalBranch{
		abstractBranch: &abstractBranch{
			oid:       oid,
			name:      name,
			shorthand: shorthand,
		},
		remoteBranch: upstreamRef,
	}

	return
}

// IsRemote returns false
func (localBranch *LocalBranch) IsRemote() bool {
	return false
}

// IsTrackingBranch returns true if this branch is tracking a remote branch
func (localBranch *LocalBranch) IsTrackingBranch() bool {
	return localBranch.remoteBranch != ""
}

// UpdateAheadBehind updates the ahead and behind counts of the branch
func (localBranch *LocalBranch) UpdateAheadBehind(ahead, behind uint) {
	localBranch.ahead = ahead
	localBranch.behind = behind
}

// Equal returns true if the other branch is a local branch equal to this one
func (localBranch *LocalBranch) Equal(other Ref) bool {
	if other == nil {
		return false
	}

	otherLocalBranch, ok := other.(*LocalBranch)
	if !ok {
		return false
	}

	return localBranch.abstractBranch.Equal(otherLocalBranch.abstractBranch) &&
		localBranch.remoteBranch == otherLocalBranch.remoteBranch
}

// RemoteBranch contains data for a remote branch reference
type RemoteBranch struct {
	*abstractBranch
}

func newRemoteBranch(oid *Oid, name, shorthand string) *RemoteBranch {
	return &RemoteBranch{
		abstractBranch: &abstractBranch{
			oid:       oid,
			name:      name,
			shorthand: shorthand,
		},
	}
}

// IsRemote returns true
func (remoteBranch *RemoteBranch) IsRemote() bool {
	return true
}

// Equal returns true if the other branch is a remote branch equal to this one
func (remoteBranch *RemoteBranch) Equal(other Ref) bool {
	if other == nil {
		return false
	}

	otherRemoteBranch, ok := other.(*RemoteBranch)
	if !ok {
		return false
	}

	return remoteBranch.abstractBranch.Equal(otherRemoteBranch.abstractBranch)
}

// Tag contains data for a tag reference
type Tag struct {
	oid       *Oid
	name      string
	shorthand string
	isRemote  bool
}

// Oid pointed to by this tag
func (tag *Tag) Oid() *Oid {
	return tag.oid
}

// Name of this tag
func (tag *Tag) Name() string {
	return tag.name
}

// Shorthand name of this tag
func (tag *Tag) Shorthand() string {
	return tag.shorthand
}

// Equal returns true if the other ref is a tag equal to this one
func (tag *Tag) Equal(other Ref) bool {
	if other == nil {
		return false
	}

	otherTag, ok := other.(*Tag)
	if !ok {
		return false
	}

	return tag.Name() == otherTag.Name() &&
		tag.Oid().Equal(otherTag.Oid())
}

// Tag returns tag data in a string format
func (tag *Tag) String() string {
	return fmt.Sprintf("%v:%v", tag.name, tag.oid)
}

// HEAD represents the HEAD ref
type HEAD struct {
	oid *Oid
}

// Oid pointed to by head
func (head *HEAD) Oid() *Oid {
	return head.oid
}

// Name of HEAD ref
func (head *HEAD) Name() string {
	return RdlHeadRef
}

// Shorthand name of HEAD ref
func (head *HEAD) Shorthand() string {
	return head.Name()
}

// IsRemote is always false
func (head *HEAD) IsRemote() bool {
	return false
}

// Equal returns true if the other ref is a HEAD equal to this one
func (head *HEAD) Equal(other Ref) bool {
	if other == nil {
		return false
	}

	otherHead, ok := other.(*HEAD)
	if !ok {
		return false
	}

	return head.Oid().Equal(otherHead.Oid())
}

// Commit contains data for a commit
type Commit struct {
	oid    *Oid
	commit *git.Commit
}

// Diff contains data for a generated diff
type Diff struct {
	diffText bytes.Buffer
	stats    bytes.Buffer
}

// StatusEntryType describes the type of change a status entry has undergone
type StatusEntryType int

// The set of supported StatusEntryTypes
const (
	SetNew StatusEntryType = iota
	SetModified
	SetDeleted
	SetRenamed
	SetTypeChange
	SetConflicted
)

var statusEntryTypeMap = map[git.Status]StatusEntryType{
	git.StatusIndexNew:        SetNew,
	git.StatusIndexModified:   SetModified,
	git.StatusIndexDeleted:    SetDeleted,
	git.StatusIndexRenamed:    SetRenamed,
	git.StatusIndexTypeChange: SetTypeChange,
	git.StatusWtNew:           SetNew,
	git.StatusWtModified:      SetModified,
	git.StatusWtDeleted:       SetDeleted,
	git.StatusWtTypeChange:    SetTypeChange,
	git.StatusWtRenamed:       SetRenamed,
	git.StatusConflicted:      SetConflicted,
}

// StatusEntry contains data for a single status entry
type StatusEntry struct {
	statusEntryType StatusEntryType
	diffDelta       git.DiffDelta
	rawStatusEntry  git.StatusEntry
}

func newStatusEntry(gitStatus git.Status, statusType StatusType, rawStatusEntry git.StatusEntry) *StatusEntry {
	var diffDelta git.DiffDelta

	if statusType == StStaged {
		diffDelta = rawStatusEntry.HeadToIndex
	} else {
		diffDelta = rawStatusEntry.IndexToWorkdir
	}

	return &StatusEntry{
		statusEntryType: statusEntryTypeMap[gitStatus],
		diffDelta:       diffDelta,
		rawStatusEntry:  rawStatusEntry,
	}
}

// StatusType describes the different stages a status entry can be in
type StatusType int

// The different status stages
const (
	StStaged StatusType = iota
	StUnstaged
	StUntracked
	StConflicted
)

var statusTypeDisplayNames = map[StatusType]string{
	StStaged:     "Staged",
	StUnstaged:   "Unstaged",
	StUntracked:  "Untracked",
	StConflicted: "Conflicted",
}

// StatusTypeDisplayName returns the display name of the StatusType
func StatusTypeDisplayName(statusType StatusType) string {
	return statusTypeDisplayNames[statusType]
}

var statusTypeMap = map[git.Status]StatusType{
	git.StatusIndexNew | git.StatusIndexModified | git.StatusIndexDeleted | git.StatusIndexRenamed | git.StatusIndexTypeChange: StStaged,
	git.StatusWtModified | git.StatusWtDeleted | git.StatusWtTypeChange | git.StatusWtRenamed:                                  StUnstaged,
	git.StatusWtNew:      StUntracked,
	git.StatusConflicted: StConflicted,
}

// Status contains all git status data
type Status struct {
	entries map[StatusType][]*StatusEntry
}

func newStatus() *Status {
	return &Status{
		entries: make(map[StatusType][]*StatusEntry),
	}
}

// StatusTypes returns the current status stages which have entries
func (status *Status) StatusTypes() (statusTypes []StatusType) {
	for statusType := range status.entries {
		statusTypes = append(statusTypes, statusType)
	}

	slice.Sort(statusTypes, func(i, j int) bool {
		return statusTypes[i] < statusTypes[j]
	})

	return
}

// Entries returns the status entries for the provided status type
func (status *Status) Entries(statusType StatusType) []*StatusEntry {
	statusEntries, ok := status.entries[statusType]
	if !ok {
		return nil
	}

	return statusEntries
}

// IsEmpty returns true if there are no entries
func (status *Status) IsEmpty() bool {
	entryNum := 0

	for _, statusEntries := range status.entries {
		entryNum += len(statusEntries)
	}

	return entryNum == 0
}

func (status *Status) addEntry(rawStatusEntry git.StatusEntry) {
	for rawStatus, statusType := range statusTypeMap {
		processedRawStatus := rawStatusEntry.Status & rawStatus

		if processedRawStatus > 0 {
			if _, ok := status.entries[statusType]; !ok {
				statusEntries := make([]*StatusEntry, 0)
				status.entries[statusType] = statusEntries
			}

			status.entries[statusType] = append(status.entries[statusType],
				newStatusEntry(processedRawStatus, statusType, rawStatusEntry))
		}
	}
}

// Equal returns true if both status' contain the same files in the same stages
func (status *Status) Equal(other *Status) bool {
	statusTypes := status.StatusTypes()
	otherStatusTypes := other.StatusTypes()

	if !reflect.DeepEqual(statusTypes, otherStatusTypes) {
		return false
	}

	for _, statusType := range statusTypes {
		if !statusEntriesEqual(status.Entries(statusType), other.Entries(statusType)) {
			return false
		}
	}

	return true
}

func statusEntriesEqual(entries, otherEntries []*StatusEntry) bool {
	if len(entries) != len(otherEntries) {
		return false
	}

	// Simply check if the same set of files have been modified in the same way
	for entryIndex, entry := range entries {
		otherEntry := otherEntries[entryIndex]

		if entry.diffDelta.Status != otherEntry.diffDelta.Status ||
			entry.diffDelta.NewFile.Path != otherEntry.diffDelta.NewFile.Path {
			return false
		}
	}

	return true
}

func newInstanceCache() *instanceCache {
	return &instanceCache{
		oids:    make(map[string]*Oid),
		commits: make(map[string]*Commit),
	}
}

func (cache *instanceCache) getOid(rawOid *git.Oid) *Oid {
	cache.oidLock.Lock()
	defer cache.oidLock.Unlock()

	oidStr := rawOid.String()

	if oid, ok := cache.oids[oidStr]; ok {
		return oid
	}

	oid := &Oid{oid: rawOid}
	cache.oids[oidStr] = oid

	return oid
}

func (cache *instanceCache) getCommit(rawCommit *git.Commit) *Commit {
	cache.commitLock.Lock()
	defer cache.commitLock.Unlock()

	oidStr := rawCommit.Id().String()

	if commit, ok := cache.commits[oidStr]; ok {
		return commit
	}

	commit := &Commit{
		oid:    cache.getOid(rawCommit.Id()),
		commit: rawCommit,
	}
	cache.commits[oidStr] = commit

	return commit
}

func (cache *instanceCache) getCachedCommit(oid *Oid) (commit *Commit, exists bool) {
	cache.commitLock.Lock()
	defer cache.commitLock.Unlock()

	commit, exists = cache.commits[oid.String()]

	return
}

func (cache *instanceCache) getCachedOid(oidStr string) (oid *Oid, exists bool) {
	cache.commitLock.Lock()
	defer cache.commitLock.Unlock()

	oid, exists = cache.oids[oidStr]

	return
}

// NewRepoDataLoader creates a new instance
func NewRepoDataLoader(channels *Channels) *RepoDataLoader {
	return &RepoDataLoader{
		cache:    newInstanceCache(),
		channels: channels,
	}
}

// Initialise attempts to access the repository
func (repoDataLoader *RepoDataLoader) Initialise(repoSupplier RepoSupplier) {
	repoDataLoader.repo = repoSupplier.RepositoryInstance()
}

// Path returns the file path location of the repository
func (repoDataLoader *RepoDataLoader) Path() string {
	return repoDataLoader.repo.Path()
}

// Head loads the current HEAD ref
func (repoDataLoader *RepoDataLoader) Head() (ref Ref, err error) {
	log.Debug("Loading HEAD")
	rawRef, err := repoDataLoader.repo.Head()
	if err != nil {
		return
	}

	oid := repoDataLoader.cache.getOid(rawRef.Target())

	if rawRef.IsBranch() {
		rawBranch := rawRef.Branch()
		ref, err = newLocalBranch(oid, rawBranch)

		if err != nil {
			log.Debugf("Failed to create branch ref for HEAD: %v", err)
			return
		}
	} else {
		ref = &HEAD{
			oid: oid,
		}
	}

	log.Debugf("Loaded HEAD %v", oid)

	return
}

// LoadRefs loads all branches and tags present in the repository
func (repoDataLoader *RepoDataLoader) LoadRefs() (refs []Ref, err error) {
	branches, err := repoDataLoader.loadBranches()
	if err != nil {
		return
	}

	tags, err := repoDataLoader.loadTags()
	if err != nil {
		return
	}

	head, err := repoDataLoader.Head()
	if err != nil {
		return
	}

	if _, isDetached := head.(*HEAD); isDetached {
		refs = append(refs, head)
	}

	for _, branch := range branches {
		refs = append(refs, branch)
	}

	for _, tag := range tags {
		refs = append(refs, tag)
	}

	return
}

func (repoDataLoader *RepoDataLoader) loadBranches() (branches []Branch, err error) {
	branchIter, err := repoDataLoader.repo.NewBranchIterator(git.BranchAll)
	if err != nil {
		return
	}
	defer branchIter.Free()

	err = branchIter.ForEach(func(branch *git.Branch, branchType git.BranchType) error {
		if repoDataLoader.channels.Exit() {
			return errors.New("Program exiting - Aborting loading branches")
		}

		branchName, err := branch.Name()
		if err != nil {
			return err
		}

		rawOid := branch.Target()

		if rawOid == nil {
			ref, err := branch.Resolve()
			if err != nil {
				return err
			}

			rawOid = ref.Target()
		}

		oid := repoDataLoader.cache.getOid(rawOid)
		var newBranch Branch

		if branch.IsRemote() {
			newBranch = newRemoteBranch(oid, branch.Reference.Name(), branchName)
		} else {
			newBranch, err = newLocalBranch(oid, branch)
			if err != nil {
				log.Debugf("Failed to create ref instance for branch %v: %v",
					branch.Reference.Name(), err)
				return err
			}
		}

		branches = append(branches, newBranch)
		log.Debugf("Loaded branch %v", newBranch)

		return nil
	})

	return
}

func (repoDataLoader *RepoDataLoader) loadTags() (tags []*Tag, err error) {
	log.Debug("Loading local tags")

	refIter, err := repoDataLoader.repo.NewReferenceIterator()
	if err != nil {
		return
	}
	defer refIter.Free()

	for {
		ref, err := refIter.Next()
		if err != nil || repoDataLoader.channels.Exit() {
			break
		}

		if !ref.IsRemote() && ref.IsTag() {
			oid := repoDataLoader.cache.getOid(ref.Target())

			newTag := &Tag{
				oid:       oid,
				name:      ref.Name(),
				shorthand: ref.Shorthand(),
			}
			tags = append(tags, newTag)

			log.Debugf("Loaded tag %v", newTag)
		}
	}

	return
}

// Commits loads all commits for the provided ref and returns a channel from which the loaded commits can be read
func (repoDataLoader *RepoDataLoader) Commits(oid *Oid) (<-chan *Commit, error) {
	revWalk, err := repoDataLoader.repo.Walk()
	if err != nil {
		return nil, err
	}

	if err := revWalk.Push(oid.oid); err != nil {
		return nil, err
	}

	log.Debugf("Loading commits for oid %v", oid)

	return repoDataLoader.loadCommits(revWalk), nil
}

// CommitRange accepts a range of the form rev..rev and returns a stream of commits in this range
func (repoDataLoader *RepoDataLoader) CommitRange(commitRange string) (<-chan *Commit, error) {
	revWalk, err := repoDataLoader.repo.Walk()
	if err != nil {
		return nil, err
	}

	if err := revWalk.PushRange(commitRange); err != nil {
		return nil, err
	}

	log.Debugf("Loading commits for range %v", commitRange)

	return repoDataLoader.loadCommits(revWalk), nil
}

func (repoDataLoader *RepoDataLoader) loadCommits(revWalk *git.RevWalk) <-chan *Commit {
	commitCh := make(chan *Commit, rdlCommitBufferSize)

	go func() {
		defer close(commitCh)
		defer revWalk.Free()

		commitNum := 0

		if err := revWalk.Iterate(func(commit *git.Commit) bool {
			if repoDataLoader.channels.Exit() {
				return false
			}

			commitNum++
			commitCh <- repoDataLoader.cache.getCommit(commit)

			return true
		}); err != nil {
			log.Errorf("Error when iterating over commits: %v", err)
		}

		log.Debugf("Loaded %v commits", commitNum)
	}()

	return commitCh
}

// Commit loads a commit for the provided oid (if it points to a commit)
func (repoDataLoader *RepoDataLoader) Commit(oid *Oid) (commit *Commit, err error) {
	if cachedCommit, isCached := repoDataLoader.cache.getCachedCommit(oid); isCached {
		return cachedCommit, nil
	}

	object, err := repoDataLoader.repo.Lookup(oid.oid)
	if err != nil {
		log.Debugf("Error when attempting to lookup object with ID %v", oid)
		return
	}

	var rawCommit *git.Commit

	switch object.Type() {
	case git.ObjectCommit:
		rawCommit, err = object.AsCommit()
		if err != nil {
			log.Debugf("Error when attempting convert object with ID %v to commit", oid)
			return
		}
	case git.ObjectTag:
		var tag *git.Tag
		tag, err = object.AsTag()
		if err != nil {
			log.Debugf("Error when attempting convert object with ID %v to tag", oid)
			return
		}

		if tag.TargetType() != git.ObjectCommit {
			err = fmt.Errorf("Tag with ID %v does not point to a commit", oid)
			return
		}

		rawCommit, err = tag.Target().AsCommit()
		if err != nil {
			log.Debugf("Error when attempting convert tag with ID %v to commit", oid)
			return
		}
	default:
		log.Debugf("Unable to convert object with type %v and ID %v to a commit", object.Type().String(), oid)
		return
	}

	commit = repoDataLoader.cache.getCommit(rawCommit)

	return
}

// CommitByOid loads a commit for the provided oid string (if it points to a commit)
func (repoDataLoader *RepoDataLoader) CommitByOid(oidStr string) (*Commit, error) {
	oid, exists := repoDataLoader.cache.getCachedOid(oidStr)
	if !exists {
		rawOid, err := git.NewOid(oidStr)
		if err != nil {
			return nil, err
		}

		oid = &Oid{oid: rawOid}
	}

	return repoDataLoader.Commit(oid)
}

// MergeBase finds the best common ancestor between two commits
func (repoDataLoader *RepoDataLoader) MergeBase(oid1, oid2 *Oid) (commonAncestor *Oid, err error) {
	rawOid, err := repoDataLoader.repo.MergeBase(oid1.oid, oid2.oid)
	if err != nil {
		err = fmt.Errorf("Unable to find common ancestor for oids %v and %v: %v", oid1, oid2, err)
	}

	commonAncestor = repoDataLoader.cache.getOid(rawOid)

	return
}

// AheadBehind returns the number of unique commits between two branches
func (repoDataLoader *RepoDataLoader) AheadBehind(local, upstream *Oid) (ahead, behind int, err error) {
	return repoDataLoader.repo.AheadBehind(local.oid, upstream.oid)
}

// DiffCommit loads a diff between the commit with the specified oid and its parent
// If the commit has more than one parent no diff is returned
func (repoDataLoader *RepoDataLoader) DiffCommit(commit *Commit) (diff *Diff, err error) {
	diff = &Diff{}

	if commit.commit.ParentCount() > 1 {
		return
	}

	var commitTree, parentTree *git.Tree
	if commitTree, err = commit.commit.Tree(); err != nil {
		return
	}
	defer commitTree.Free()

	if commit.commit.ParentCount() > 0 {
		if parentTree, err = commit.commit.Parent(0).Tree(); err != nil {
			return
		}
		defer parentTree.Free()
	}

	options, err := git.DefaultDiffOptions()
	if err != nil {
		return
	}

	commitDiff, err := repoDataLoader.repo.DiffTreeToTree(parentTree, commitTree, &options)
	if err != nil {
		return
	}
	defer commitDiff.Free()

	return repoDataLoader.generateDiff(commitDiff)
}

// DiffStage returns a diff for all files in the provided stage
func (repoDataLoader *RepoDataLoader) DiffStage(statusType StatusType) (diff *Diff, err error) {
	diff = &Diff{}

	rawDiff, err := repoDataLoader.generateRawDiff(statusType)
	if err != nil || rawDiff == nil {
		return
	}
	defer rawDiff.Free()

	return repoDataLoader.generateDiff(rawDiff)
}

// DiffFile Generates a diff for the provided file
// If statusType is StStaged then the diff is between HEAD and the index
// If statusType is StUnstaged then the diff is between index and the working directory
func (repoDataLoader *RepoDataLoader) DiffFile(statusType StatusType, path string) (diff *Diff, err error) {
	diff = &Diff{}

	rawDiff, err := repoDataLoader.generateRawDiff(statusType)
	if err != nil || rawDiff == nil {
		return
	}
	defer rawDiff.Free()

	numDeltas, err := rawDiff.NumDeltas()
	if err != nil {
		return
	}

	var diffDelta git.DiffDelta
	var patch *git.Patch
	var patchString string

	for i := 0; i < numDeltas; i++ {
		if diffDelta, err = rawDiff.GetDelta(i); err != nil {
			return
		}

		if diffDelta.NewFile.Path == path {
			if patch, err = rawDiff.Patch(i); err != nil {
				return
			}

			if patchString, err = patch.String(); err != nil {
				return
			}

			diff.diffText.WriteString(patchString)

			if err := patch.Free(); err != nil {
				log.Errorf("Error when freeing patch: %v", err)
			}

			break
		}
	}

	return
}

func (repoDataLoader *RepoDataLoader) generateRawDiff(statusType StatusType) (rawDiff *git.Diff, err error) {
	var index *git.Index
	var options git.DiffOptions

	switch statusType {
	case StStaged:
		var head Ref
		var commit *Commit
		var tree *git.Tree

		if head, err = repoDataLoader.Head(); err != nil {
			return
		}

		if commit, err = repoDataLoader.Commit(head.Oid()); err != nil {
			return
		}

		if tree, err = commit.commit.Tree(); err != nil {
			return
		}

		if index, err = repoDataLoader.repo.Index(); err != nil {
			return
		}

		if options, err = git.DefaultDiffOptions(); err != nil {
			return
		}

		if rawDiff, err = repoDataLoader.repo.DiffTreeToIndex(tree, index, &options); err != nil {
			return
		}
	case StUnstaged:
		if index, err = repoDataLoader.repo.Index(); err != nil {
			return
		}

		if options, err = git.DefaultDiffOptions(); err != nil {
			return
		}

		if rawDiff, err = repoDataLoader.repo.DiffIndexToWorkdir(index, &options); err != nil {
			return
		}
	}

	return
}

func (repoDataLoader *RepoDataLoader) generateDiff(rawDiff *git.Diff) (diff *Diff, err error) {
	diff = &Diff{}

	stats, err := rawDiff.Stats()
	if err != nil {
		return
	}

	statsText, err := stats.String(git.DiffStatsFull, rdlDiffStatsCols)
	if err != nil {
		return
	}

	diff.stats.WriteString(statsText)

	numDeltas, err := rawDiff.NumDeltas()
	if err != nil {
		return
	}

	var patch *git.Patch
	var patchString string

	for i := 0; i < numDeltas; i++ {
		if patch, err = rawDiff.Patch(i); err != nil {
			return
		}

		if patchString, err = patch.String(); err != nil {
			return
		}

		diff.diffText.WriteString(patchString)

		if err := patch.Free(); err != nil {
			log.Errorf("Error when freeing patch: %v", err)
		}
	}

	return
}

// LoadStatus loads git status and populates a Status instance with the data
func (repoDataLoader *RepoDataLoader) LoadStatus() (*Status, error) {
	log.Debug("Loading git status")

	statusOptions := git.StatusOptions{
		Show:  git.StatusShowIndexAndWorkdir,
		Flags: git.StatusOptIncludeUntracked,
	}

	statusList, err := repoDataLoader.repo.StatusList(&statusOptions)
	if err != nil {
		return nil, fmt.Errorf("Unable to determine repository status: %v", err)
	}

	defer statusList.Free()

	entryCount, err := statusList.EntryCount()
	if err != nil {
		return nil, fmt.Errorf("Unable to determine repository status: %v", err)
	}

	status := newStatus()

	for i := 0; i < entryCount; i++ {
		statusEntry, err := statusList.ByIndex(i)
		if err != nil {
			return nil, fmt.Errorf("Unable to determine repository status: %v", err)
		}

		status.addEntry(statusEntry)
	}

	return status, nil
}
