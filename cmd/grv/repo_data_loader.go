package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	slice "github.com/bradfitz/slice"
	git "gopkg.in/libgit2/git2go.v27"
)

const (
	// RdlHeadRef is the HEAD ref name
	RdlHeadRef                       = "HEAD"
	rdlCommitBufferSize              = 100
	rdlDiffStatsCols                 = 80
	rdlShortOidLen                   = 7
	rdlCommitLimitDateFormat         = "2006-01-02"
	rdlCommitLimitDateTimeFormat     = "2006-01-02 15:04:05"
	rdlCommitLimitDateTimeZoneFormat = "2006-01-02 15:04:05-0700"
	// GitRepositoryDirectoryName is the name of the git directory in a git repository
	GitRepositoryDirectoryName = ".git"
)

var diffErrorRegex = regexp.MustCompile(`Invalid (regexp|collation character)`)

var noCommitLimit = regexp.MustCompile(`^\s*$`)
var numericCommitLimit = regexp.MustCompile(`^\d+$`)
var dateCommitLimit = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
var dateTimeCommitLimit = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}$`)
var dateTimeZoneCommitLimit = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}(\+|-)\d{4}$`)
var oidCommitLimit = regexp.MustCompile(`^[[:xdigit:]]+$`)

type commitLimitPredicate func(*git.Commit) bool

var noCommitLimitPredicate = func(*git.Commit) bool {
	return false
}

type instanceCache struct {
	oids       map[string]*Oid
	commits    map[string]*Commit
	oidLock    sync.Mutex
	commitLock sync.Mutex
}

// RepoDataLoader handles loading data from the repository
type RepoDataLoader struct {
	repo               *git.Repository
	cache              *instanceCache
	channels           Channels
	config             Config
	commitLimitReached commitLimitPredicate
	diffErrorPresent   bool
	gitBinaryConfirmed bool
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
	remoteName string
}

func newRemoteBranch(oid *Oid, remoteName, name, shorthand string) *RemoteBranch {
	return &RemoteBranch{
		abstractBranch: &abstractBranch{
			oid:       oid,
			name:      name,
			shorthand: shorthand,
		},
		remoteName: remoteName,
	}
}

// ShorthandWithoutRemote returns only the branch name
func (remoteBranch *RemoteBranch) ShorthandWithoutRemote() string {
	return strings.TrimPrefix(remoteBranch.shorthand, remoteBranch.remoteName+"/")
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

// String returns tag data in a string format
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

// String returns HEAD in a string format
func (head *HEAD) String() string {
	return fmt.Sprintf("%v:%v", head.Name(), head.Oid())
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

// NewFilePath returns the new file path of the status entry
func (statusEntry *StatusEntry) NewFilePath() string {
	return statusEntry.diffDelta.NewFile.Path
}

// OldFilePath returns the old file path of the status entry
func (statusEntry *StatusEntry) OldFilePath() string {
	return statusEntry.diffDelta.OldFile.Path
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

// RepositoryState describes the state of the repository
// e.g. whether an operation is in progress
type RepositoryState int

// Set of supported repository states
const (
	RepositoryStateUnknown RepositoryState = iota
	RepositoryStateNone
	RepositoryStateMerge
	RepositoryStateRevert
	RepositoryStateCherrypick
	RepositoryStateBisect
	RepositoryStateRebase
	RepositoryStateRebaseInteractive
	RepositoryStateRebaseMerge
	RepositoryStateApplyMailbox
	RepositoryStateApplyMailboxOrRebase
)

var repositoryStateMap = map[git.RepositoryState]RepositoryState{
	git.RepositoryStateNone:                 RepositoryStateNone,
	git.RepositoryStateMerge:                RepositoryStateMerge,
	git.RepositoryStateRevert:               RepositoryStateRevert,
	git.RepositoryStateCherrypick:           RepositoryStateCherrypick,
	git.RepositoryStateBisect:               RepositoryStateBisect,
	git.RepositoryStateRebase:               RepositoryStateRebase,
	git.RepositoryStateRebaseInteractive:    RepositoryStateRebaseInteractive,
	git.RepositoryStateRebaseMerge:          RepositoryStateRebaseMerge,
	git.RepositoryStateApplyMailbox:         RepositoryStateApplyMailbox,
	git.RepositoryStateApplyMailboxOrRebase: RepositoryStateApplyMailboxOrRebase,
}

// Status contains all git status data
type Status struct {
	repositoryState RepositoryState
	entries         map[StatusType][]*StatusEntry
}

func newStatus(repositoryState RepositoryState) *Status {
	return &Status{
		repositoryState: repositoryState,
		entries:         make(map[StatusType][]*StatusEntry),
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
func (status *Status) Entries(statusType StatusType) (statusEntries []*StatusEntry) {
	statusEntries, ok := status.entries[statusType]
	if !ok {
		return
	}

	return statusEntries
}

// FilePaths returns the paths of all files with the provided StatusType
func (status *Status) FilePaths(statusType StatusType) (filePaths []string) {
	statusEntries, ok := status.entries[statusType]
	if !ok {
		return
	}

	for _, statusEntry := range statusEntries {
		filePaths = append(filePaths, statusEntry.NewFilePath())
	}

	return
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

// RepositoryState returns the current repository state
func (status *Status) RepositoryState() RepositoryState {
	return status.repositoryState
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

func (repoDataLoader *RepoDataLoader) newCommitLimiter(commitLimitString string) (commitLimitReached commitLimitPredicate, err error) {
	commitLimitReached = noCommitLimitPredicate

	switch {
	case noCommitLimit.MatchString(commitLimitString):
	case numericCommitLimit.MatchString(commitLimitString):
		var limit int
		if limit, err = strconv.Atoi(commitLimitString); err != nil {
			err = fmt.Errorf("Unable to parse commit limit: %v", commitLimitString)
			return
		}

		commitCount := 0
		commitLimitReached = func(*git.Commit) bool {
			if commitCount >= limit {
				return true
			}

			commitCount++
			return false
		}
	case dateCommitLimit.MatchString(commitLimitString):
		if commitLimitReached, err = generateDateCommitLimitTester(commitLimitString, rdlCommitLimitDateFormat, false); err != nil {
			err = fmt.Errorf("Failed to parse commit limit date string: %v", err)
			return
		}
	case dateTimeCommitLimit.MatchString(commitLimitString):
		if commitLimitReached, err = generateDateCommitLimitTester(commitLimitString, rdlCommitLimitDateTimeFormat, false); err != nil {
			err = fmt.Errorf("Failed to parse commit limit date-time string: %v", err)
			return
		}
	case dateTimeZoneCommitLimit.MatchString(commitLimitString):
		if commitLimitReached, err = generateDateCommitLimitTester(commitLimitString, rdlCommitLimitDateTimeZoneFormat, true); err != nil {
			err = fmt.Errorf("Failed to parse commit limit date-time string: %v", err)
			return
		}
	case oidCommitLimit.MatchString(commitLimitString):
		var object *git.Object
		if object, err = repoDataLoader.repo.RevparseSingle(commitLimitString); err != nil {
			err = fmt.Errorf("Invalid oid for commit limit: %v", err)
			return
		}
		defer object.Free()

		if object.Type() != git.ObjectCommit {
			err = fmt.Errorf("Oid for commit limit does not reference commit: %v", commitLimitString)
			return
		}

		oid := object.Id().String()

		commitSeen := false
		commitLimitReached = func(commit *git.Commit) bool {
			if commitSeen {
				return true
			}

			commitSeen = commit.Id().String() == oid

			return false
		}
	default:
		var object *git.Object
		if object, err = repoDataLoader.repo.RevparseSingle(commitLimitString); err != nil {
			err = fmt.Errorf("Invalid tag for commit limit: %v", err)
			return
		}
		defer object.Free()

		if object.Type() != git.ObjectTag {
			err = fmt.Errorf("Oid for commit limit does not reference tag: %v", commitLimitString)
			return
		}

		var tag *git.Tag
		if tag, err = object.AsTag(); err != nil {
			err = fmt.Errorf("Unable to load tag with name %v for commit limit: %v", commitLimitString, err)
			return
		}

		if tag.TargetType() != git.ObjectCommit {
			err = fmt.Errorf("Tag for commit limit does not reference commit: %v", commitLimitString)
			return
		}

		var commit *git.Commit
		if commit, err = tag.Target().AsCommit(); err != nil {
			err = fmt.Errorf("Unable to load commit for commit limit tag %v: %v", commitLimitString, err)
			return
		}

		oid := commit.Id().String()

		commitSeen := false
		commitLimitReached = func(commit *git.Commit) bool {
			if commitSeen {
				return true
			}

			commitSeen = commit.Id().String() == oid

			return false
		}
	}

	return
}

func generateDateCommitLimitTester(dateString string, dateFormat string, timeZone bool) (commitLimitReached commitLimitPredicate, err error) {
	commitLimitReached = noCommitLimitPredicate

	dateTime, err := time.Parse(dateFormat, dateString)
	if err != nil {
		return
	}

	if !timeZone {
		dateTime = TimeWithLocation(dateTime, time.Local)
	}

	commitLimitReached = func(commit *git.Commit) bool {
		return commit.Author().When.Before(dateTime)
	}

	return
}

// NewRepoDataLoader creates a new instance
func NewRepoDataLoader(channels Channels, config Config) *RepoDataLoader {
	return &RepoDataLoader{
		cache:            newInstanceCache(),
		channels:         channels,
		config:           config,
		diffErrorPresent: true,
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

// Workdir returns working directory file path for the repository
func (repoDataLoader *RepoDataLoader) Workdir() string {
	return repoDataLoader.repo.Workdir()
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
			fullBranchName := branch.Reference.Name()
			remoteName, err := repoDataLoader.repo.RemoteName(fullBranchName)
			if err != nil {
				err = fmt.Errorf("Failed to determine remote for branch %v: %v", fullBranchName, err)
				return err
			}

			newBranch = newRemoteBranch(oid, remoteName, fullBranchName, branchName)
		} else {
			newBranch, err = newLocalBranch(oid, branch)
			if err != nil {
				err = fmt.Errorf("Failed to create ref instance for branch %v: %v",
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

	var ref *git.Reference

	for {
		if repoDataLoader.channels.Exit() {
			break
		}

		if ref, err = refIter.Next(); err != nil {
			if gitError, isGitError := err.(*git.GitError); !isGitError || gitError.Code != git.ErrIterOver {
				err = fmt.Errorf("Error when loading tags: %v", err)
			} else {
				err = nil
			}

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
	commitLimit := repoDataLoader.config.GetString(CfCommitLimit)

	commitLimitReached, err := repoDataLoader.newCommitLimiter(commitLimit)
	if err != nil {
		repoDataLoader.channels.ReportError(err)
	}

	go func() {
		defer close(commitCh)
		defer revWalk.Free()

		commitNum := 0

		if err := revWalk.Iterate(func(commit *git.Commit) bool {
			if repoDataLoader.channels.Exit() {
				return false
			} else if commitLimitReached(commit) {
				repoDataLoader.channels.ReportStatus("Commit limit reached")
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

	if repoDataLoader.diffErrorPresent {
		return repoDataLoader.generateCommitDiffUsingCLI(commit)
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

	if diff, err = repoDataLoader.generateDiff(commitDiff); err != nil && diffErrorRegex.MatchString(err.Error()) {
		log.Infof("Falling back to git cli after encountering error: %v", err)
		repoDataLoader.diffErrorPresent = true
		return repoDataLoader.generateCommitDiffUsingCLI(commit)
	}

	return
}

// DiffStage returns a diff for all files in the provided stage
func (repoDataLoader *RepoDataLoader) DiffStage(statusType StatusType) (diff *Diff, err error) {
	if repoDataLoader.diffErrorPresent {
		return repoDataLoader.generateStageDiffUsingCLI(statusType)
	}

	diff = &Diff{}

	rawDiff, err := repoDataLoader.generateRawDiff(statusType)
	if err != nil {
		if diffErrorRegex.MatchString(err.Error()) {
			log.Infof("Falling back to git cli after encountering error: %v", err)
			repoDataLoader.diffErrorPresent = true
			return repoDataLoader.generateStageDiffUsingCLI(statusType)
		}

		return
	} else if rawDiff == nil {
		err = fmt.Errorf("Failed to generate diff for %v files", StatusTypeDisplayName(statusType))
		return
	}
	defer rawDiff.Free()

	return repoDataLoader.generateDiff(rawDiff)
}

// DiffFile Generates a diff for the provided file
// If statusType is StStaged then the diff is between HEAD and the index
// If statusType is StUnstaged then the diff is between index and the working directory
func (repoDataLoader *RepoDataLoader) DiffFile(statusType StatusType, path string) (diff *Diff, err error) {
	if repoDataLoader.diffErrorPresent {
		return repoDataLoader.generateFileDiffUsingCLI(statusType, path)
	}

	diff = &Diff{}

	rawDiff, err := repoDataLoader.generateRawDiff(statusType)
	if err != nil {
		if diffErrorRegex.MatchString(err.Error()) {
			log.Infof("Falling back to git cli after encountering error: %v", err)
			repoDataLoader.diffErrorPresent = true
			return repoDataLoader.generateFileDiffUsingCLI(statusType, path)
		}

		return
	} else if rawDiff == nil {
		err = fmt.Errorf("Failed to generate diff for %v file %v", StatusTypeDisplayName(statusType), path)
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
	var head Ref
	var commit *Commit
	var tree *git.Tree

	switch statusType {
	case StStaged:
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
	case StConflicted:
		if head, err = repoDataLoader.Head(); err != nil {
			return
		}

		if commit, err = repoDataLoader.Commit(head.Oid()); err != nil {
			return
		}

		if tree, err = commit.commit.Tree(); err != nil {
			return
		}

		if options, err = git.DefaultDiffOptions(); err != nil {
			return
		}

		if rawDiff, err = repoDataLoader.repo.DiffTreeToWorkdir(tree, &options); err != nil {
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

type diffType int

const (
	dtCommit diffType = iota
	dtStage
	dtFile
)

func (repoDataLoader *RepoDataLoader) generateCommitDiffUsingCLI(commit *Commit) (diff *Diff, err error) {
	log.Debugf("Attempting to load diff using cli for commit: %v", commit.oid.String())
	gitCommand := []string{"show", "--encoding=UTF8", "--pretty=oneline", "--root", "--patch-with-stat", "--no-color", commit.oid.String()}
	return repoDataLoader.runGitCLIDiff(gitCommand, dtCommit)
}

func (repoDataLoader *RepoDataLoader) generateFileDiffUsingCLI(statusType StatusType, path string) (diff *Diff, err error) {
	log.Debugf("Attempting to load diff using cli for StatusType: %v and file: %v", StatusTypeDisplayName(statusType), path)

	gitCommand := []string{"diff"}

	if statusType == StStaged {
		gitCommand = append(gitCommand, "--cached")
	} else if statusType != StUnstaged {
		return &Diff{}, nil
	}

	gitCommand = append(gitCommand, []string{"--encoding=UTF8", "--root", "--no-color", "--", path}...)

	return repoDataLoader.runGitCLIDiff(gitCommand, dtFile)
}

func (repoDataLoader *RepoDataLoader) generateStageDiffUsingCLI(statusType StatusType) (diff *Diff, err error) {
	log.Debugf("Attempting to load diff using cli for StatusType: %v", StatusTypeDisplayName(statusType))

	gitCommand := []string{"diff"}

	if statusType == StStaged {
		gitCommand = append(gitCommand, "--cached")
	} else if statusType != StUnstaged {
		return &Diff{}, nil
	}

	gitCommand = append(gitCommand, []string{"--encoding=UTF8", "--root", "--patch-with-stat", "--no-color"}...)

	return repoDataLoader.runGitCLIDiff(gitCommand, dtStage)
}

func (repoDataLoader *RepoDataLoader) gitBinary() string {
	if gitBinary := repoDataLoader.config.GetString(CfGitBinaryFilePath); gitBinary != "" {
		return gitBinary
	}

	return "git"
}

func (repoDataLoader *RepoDataLoader) runGitCLIDiff(gitCommand []string, diffType diffType) (diff *Diff, err error) {
	diff = &Diff{}

	if !repoDataLoader.gitBinaryConfirmed {
		if exec.Command(repoDataLoader.gitBinary(), "version").Run() == nil {
			repoDataLoader.gitBinaryConfirmed = true
		} else {
			err = fmt.Errorf("Unable to successfully call git binary. "+
				"If git is not in $PATH then please set the config variable %v", CfGitBinaryFilePath)
			return
		}
	}

	cmd := exec.Command(repoDataLoader.gitBinary(), gitCommand...)
	cmd.Env, cmd.Dir = repoDataLoader.GenerateGitCommandEnvironment()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		err = fmt.Errorf("Unable to generate commit diff using git cli: %v", err)
		return
	}

	if stderr.Len() > 0 {
		err = fmt.Errorf("Error when generating commit diff using git cli: %v", stderr.String())
		return
	}

	scanner := bufio.NewScanner(&stdout)

	if diffType != dtFile {
		if diffType == dtCommit {
			scanner.Scan()
		}

		for scanner.Scan() {
			if line := scanner.Text(); line == "" {
				break
			} else {
				diff.stats.WriteString(strings.TrimPrefix(line, " "))
				diff.stats.WriteRune('\n')
			}
		}
	}

	for scanner.Scan() {
		diff.diffText.WriteString(scanner.Text())
		diff.diffText.WriteRune('\n')
	}

	if err = scanner.Err(); err != nil {
		err = fmt.Errorf("Reading commit diff cli output failed: %v", err)
		return
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

	repositoryState := repoDataLoader.RepositoryState()
	status := newStatus(repositoryState)

	for i := 0; i < entryCount; i++ {
		statusEntry, err := statusList.ByIndex(i)
		if err != nil {
			return nil, fmt.Errorf("Unable to determine repository status: %v", err)
		}

		status.addEntry(statusEntry)
	}

	return status, nil
}

// UserEditor returns the editor git is configured to use
func (repoDataLoader *RepoDataLoader) UserEditor() (editor string, err error) {
	config, err := repoDataLoader.repo.Config()
	if err != nil {
		err = fmt.Errorf("Unable to retrieve git config: %v", err)
	}

	if editor, _ = config.LookupString("core.editor"); editor != "" {
		return
	}

	editor = os.Getenv("GIT_EDITOR")
	return
}

// GenerateGitCommandEnvironment populates git environment variables for
// the current repository
func (repoDataLoader *RepoDataLoader) GenerateGitCommandEnvironment() (env []string, rootDir string) {
	env = os.Environ()

	gitDir := repoDataLoader.Path()
	if gitDir != "" {
		env = append(env, fmt.Sprintf("GIT_DIR=%v", gitDir))
	}

	workdir := repoDataLoader.Workdir()
	if workdir != "" {
		env = append(env, fmt.Sprintf("GIT_WORK_TREE=%v", workdir))
		rootDir = workdir
	} else {
		rootDir = strings.TrimSuffix(gitDir, GitRepositoryDirectoryName)
	}

	return
}

// RepositoryState returns the current repository state
func (repoDataLoader *RepoDataLoader) RepositoryState() RepositoryState {
	repositoryState := repoDataLoader.repo.State()

	if mappedRepositoryState, ok := repositoryStateMap[repositoryState]; ok {
		return mappedRepositoryState
	}

	return RepositoryStateUnknown
}

// Remotes loads remotes for the repository
func (repoDataLoader *RepoDataLoader) Remotes() (remotes []string, err error) {
	if remotes, err = repoDataLoader.repo.Remotes.List(); err != nil {
		err = fmt.Errorf("Failed to determine remotes: %v", err)
	}

	return
}
