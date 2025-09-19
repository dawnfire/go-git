# patch github.com/go-git/go-git/v5@5.16.2

## 1 how to PATCH next version 

### 1.1 clone github.com/go-git/go-git

### 1.2 overwrite current version

### 1.3 using diff to generate bellow files differ

### 1.4 patch those files

### 1.5 optionally running the go test 

go test -v _examples/common_test.go _examples/common.go --examples




## 2 plumbing/reference.go

```diff
diff --git a/plumbing/reference.go b/plumbing/reference.go
index eef11e82..30d7d385 100644
--- a/plumbing/reference.go
+++ b/plumbing/reference.go
@@ -13,6 +13,8 @@ const (
refRemotePrefix = refPrefix + "remotes/"
refNotePrefix   = refPrefix + "notes/"
symrefPrefix    = "ref: "
+
+	logPrefix = "logs/"
     )

// RefRevParseRules are a set of rules to parse references into short names.
@@ -24,6 +26,7 @@ var RefRevParseRules = []string{
"refs/heads/%s",
"refs/remotes/%s",
"refs/remotes/%s/HEAD",
+	"logs/%s",
     }

var (
@@ -37,6 +40,8 @@ const (
InvalidReference  ReferenceType = 0
HashReference     ReferenceType = 1
SymbolicReference ReferenceType = 2
+
+	LogReference ReferenceType = 3
     )

func (r ReferenceType) String() string {
@@ -47,6 +52,8 @@ func (r ReferenceType) String() string {
return "hash-reference"
case SymbolicReference:
return "symbolic-reference"
+	case LogReference:
+		return "log-reference"
     }

     return ""
     @@ -85,6 +92,12 @@ func NewTagReferenceName(name string) ReferenceName {
     return ReferenceName(refTagPrefix + name)
     }

+// NewLogReferenceName returns a reference name describing a tag based on short
+// his name.
+func NewLogReferenceName(name string) ReferenceName {
+	return ReferenceName(logPrefix + name)
     +}
+
// IsBranch check if a reference is a branch
func (r ReferenceName) IsBranch() bool {
return strings.HasPrefix(string(r), refHeadPrefix)
@@ -95,6 +108,11 @@ func (r ReferenceName) IsNote() bool {
return strings.HasPrefix(string(r), refNotePrefix)
}

+// IsLog check if a reference is a note
+func (r ReferenceName) IsLog() bool {
+	return strings.HasPrefix(string(r), logPrefix)
     +}
+
// IsRemote check if a reference is a remote
func (r ReferenceName) IsRemote() bool {
return strings.HasPrefix(string(r), refRemotePrefix)
@@ -126,6 +144,7 @@ func (r ReferenceName) Short() string {
const (
HEAD   ReferenceName = "HEAD"
Master ReferenceName = "refs/heads/master"
+	LOG    ReferenceName = "logs"
     )

// Reference is a representation of git reference
@@ -147,6 +166,10 @@ func NewReferenceFromStrings(name, target string) *Reference {
return NewSymbolicReference(n, target)
}

+	if strings.HasPrefix(name, "logs/") {
+		return NewSymbolicReference(ReferenceName(name), ReferenceName(target))
+	}
+
return NewHashReference(n, NewHash(target))
}

@@ -159,6 +182,15 @@ func NewSymbolicReference(n, target ReferenceName) *Reference {
}
}

+// NewLogReference creates a new LogReference reference
+func NewLogReference(n, target ReferenceName) *Reference {
+	return &Reference{
+		t:      LogReference,
+		n:      n,
+		target: target,
+	}
     +}
+
// NewHashReference creates a new HashReference reference
func NewHashReference(n ReferenceName, h Hash) *Reference {
return &Reference{
@@ -210,6 +242,8 @@ func (r *Reference) String() string {
ref = r.Hash().String()
case SymbolicReference:
ref = symrefPrefix + r.Target().String()
+	case LogReference:
+		ref = r.Target().String()
     default:
     return ""
     }
```


## 3 plumbing/storer/reference.go

```diff
diff --git a/plumbing/storer/reference.go b/plumbing/storer/reference.go
index 1d74ef3c..a6b5ee7f 100644
--- a/plumbing/storer/reference.go
+++ b/plumbing/storer/reference.go
@@ -15,6 +15,8 @@ var ErrMaxResolveRecursion = errors.New("max. recursion level reached")

// ReferenceStorer is a generic storage of references.
type ReferenceStorer interface {
+	SetLog(*plumbing.Reference) error
+
SetReference(*plumbing.Reference) error
// CheckAndSetReference sets the reference `new`, but if `old` is
// not `nil`, it first checks that the current stored value for
@@ -22,7 +24,10 @@ type ReferenceStorer interface {
// not, it returns an error and doesn't update `new`.
CheckAndSetReference(new, old *plumbing.Reference) error
Reference(plumbing.ReferenceName) (*plumbing.Reference, error)
+	RefLog(refName plumbing.ReferenceName, resolved bool) (*plumbing.Reference, error)
+
IterReferences() (ReferenceIter, error)
+
RemoveReference(plumbing.ReferenceName) error
CountLooseRefs() (int, error)
PackRefs() error
```

## 4 plumbing/transport/internal/common/common.go

```diff
diff --git a/plumbing/transport/internal/common/common.go b/plumbing/transport/internal/common/common.go
index d0e9a297..b2c2fee3 100644
--- a/plumbing/transport/internal/common/common.go
+++ b/plumbing/transport/internal/common/common.go
@@ -374,7 +374,7 @@ func (s *session) checkNotFoundError() error {
case <-t.C:
return ErrTimeoutExceeded
case line, ok := <-s.firstErrLine:
-		if !ok {
+		if !ok || len(line) == 0 {
     		return nil
     	}
```

## 5 storage/filesystem/reference.go

```diff
diff --git a/storage/filesystem/reference.go b/storage/filesystem/reference.go
index aabcd730..f980b9d5 100644
--- a/storage/filesystem/reference.go
+++ b/storage/filesystem/reference.go
@@ -10,6 +10,10 @@ type ReferenceStorage struct {
dir *dotgit.DotGit
}

+func (r *ReferenceStorage) SetLog(ref *plumbing.Reference) error {
+	return r.dir.SetLog(ref, nil)
     +}
+
func (r *ReferenceStorage) SetReference(ref *plumbing.Reference) error {
return r.dir.SetRef(ref, nil)
}
@@ -22,6 +26,10 @@ func (r *ReferenceStorage) Reference(n plumbing.ReferenceName) (*plumbing.Refere
return r.dir.Ref(n)
}

+func (r *ReferenceStorage) RefLog(n plumbing.ReferenceName, resolved bool) (*plumbing.Reference, error) {
+	return r.dir.RefLog(n, resolved)
     +}
+
func (r *ReferenceStorage) IterReferences() (storer.ReferenceIter, error) {
refs, err := r.dir.Refs()
if err != nil {
```

## 6 add new file storage/filesystem/dotgit/dotgit_setlog.go

```go
package dotgit

import (
"fmt"
"io"
"os"

"github.com/go-git/go-git/v5/plumbing"
"github.com/go-git/go-git/v5/utils/ioutil"

"github.com/go-git/go-billy/v5"
)

func (d *DotGit) setLog(fileName, content string, old *plumbing.Reference) (err error) {
if billy.CapabilityCheck(d.fs, billy.ReadAndWriteCapability) {
return d.setLogRwfs(fileName, content, old)
}

return d.setLogNorwfs(fileName, content, old)
}

func (d *DotGit) setLogRwfs(fileName, content string, old *plumbing.Reference) (err error) {
// If we are not checking an old ref, just truncate the file.
mode := os.O_RDWR | os.O_CREATE | os.O_APPEND

f, err := d.fs.OpenFile(fileName, mode, 0666)
if err != nil {
return err
}

defer ioutil.CheckClose(f, &err)

// Lock is unlocked by the deferred Close above. This is because Unlock
// does not imply a fsync and thus there would be a race between
// Unlock+Close and other concurrent writers. Adding Sync to go-billy
// could work, but this is better (and avoids superfluous syncs).
err = f.Lock()
if err != nil {
return err
}

_, err = f.Seek(0, io.SeekEnd)
if err != nil {
return err
}

_, err = f.Write([]byte(content))
if err != nil {
return err
}

_, err = f.Write([]byte("\n"))
return err
}

// There are some filesystems that don't support opening files in RDWD mode.
// In these filesystems the standard SetRef function can not be used as it
// reads the reference file to check that it's not modified before updating it.
//
// This version of the function writes the reference without extra checks
// making it compatible with these simple filesystems. This is usually not
// a problem as they should be accessed by only one process at a time.
func (d *DotGit) setLogNorwfs(fileName, content string, old *plumbing.Reference) error {
_, err := d.fs.Stat(fileName)
if err == nil && old != nil {
fRead, err := d.fs.Open(fileName)
if err != nil {
return err
}

    ref, err := d.readReferenceFrom(fRead, old.Name().String())
    fRead.Close()

    if err != nil {
      return err
    }

    if ref.Hash() != old.Hash() {
      return fmt.Errorf("reference has changed concurrently")
    }
}

f, err := d.fs.Create(fileName)
if err != nil {
return err
}

defer f.Close()

_, err = f.Write([]byte(content))
return err
}
```

## 7 storage/filesystem/dotgit/dotgit.go

```diff
diff --git a/storage/filesystem/dotgit/dotgit.go b/storage/filesystem/dotgit/dotgit.go
index 6c386f79..4942b884 100644
--- a/storage/filesystem/dotgit/dotgit.go
+++ b/storage/filesystem/dotgit/dotgit.go
@@ -563,7 +563,8 @@ func (d *DotGit) objectPath(h plumbing.Hash) string {
//
// More on git hooks found here : https://git-scm.com/docs/githooks
// More on 'quarantine'/incoming directory here:
-//     https://git-scm.com/docs/git-receive-pack
+//
+//	https://git-scm.com/docs/git-receive-pack
func (d *DotGit) incomingObjectPath(h plumbing.Hash) string {
hString := h.String()

@@ -672,9 +673,27 @@ func (d *DotGit) checkReferenceAndTruncate(f billy.File, old *plumbing.Reference
return f.Truncate(0)
}

+func (d *DotGit) SetLog(r, old *plumbing.Reference) error {
+	var content string
+	switch r.Type() {
+	case plumbing.LogReference:
+		content = r.Target().String()
+	case plumbing.SymbolicReference:
+		content = fmt.Sprintf("ref: %s\n", r.Target())
+	case plumbing.HashReference:
+		content = fmt.Sprintln(r.Hash().String())
+	}
+
+	fileName := r.Name().String()
+
+	return d.setLog(fileName, content, old)
     +}
+
func (d *DotGit) SetRef(r, old *plumbing.Reference) error {
var content string
switch r.Type() {
+	case plumbing.LogReference:
+		content = r.Target().String()
     case plumbing.SymbolicReference:
     content = fmt.Sprintf("ref: %s\n", r.Target())
     case plumbing.HashReference:
     @@ -716,6 +735,11 @@ func (d *DotGit) Ref(name plumbing.ReferenceName) (*plumbing.Reference, error) {
     return d.packedRef(name)
     }

+// RefLog returns the reference for a given reference name.
+func (d *DotGit) RefLog(name plumbing.ReferenceName, resolved bool) (*plumbing.Reference, error) {
+	return d.readRefLogFile(".", name.String(), resolved)
     +}
+
func (d *DotGit) findPackedRefsInFile(f billy.File) ([]*plumbing.Reference, error) {
s := bufio.NewScanner(f)
var refs []*plumbing.Reference
@@ -1006,6 +1030,30 @@ func (d *DotGit) readReferenceFile(path, name string) (ref *plumbing.Reference,
return d.readReferenceFrom(f, name)
}

+func (d *DotGit) readRefLogFile(path, name string, resolved bool) (ref *plumbing.Reference, err error) {
+	path = d.fs.Join(path, d.fs.Join(strings.Split(name, "/")...))
+	st, err := d.fs.Stat(path)
+	if err != nil {
+		return
+	}
+	if st.IsDir() {
+		return nil, ErrIsDir
+	}
+
+	if resolved {
+		ref = plumbing.NewSymbolicReference(plumbing.ReferenceName(name), "true")
+		return
+	}
+
+	f, err := d.fs.Open(path)
+	if err != nil {
+		return nil, err
+	}
+	defer ioutil.CheckClose(f, &err)
+
+	return d.readReferenceFrom(f, name)
     +}
+
func (d *DotGit) CountLooseRefs() (int, error) {
var refs []*plumbing.Reference
var seen = make(map[plumbing.ReferenceName]bool)
```

## 8 storage/memory/storage.go

```diff
diff --git a/storage/memory/storage.go b/storage/memory/storage.go
index ef6a4455..94191037 100644
--- a/storage/memory/storage.go
+++ b/storage/memory/storage.go
@@ -3,6 +3,7 @@ package memory

import (
"fmt"
+	"os"
     "time"

 	"github.com/go-git/go-git/v5/config"
@@ -241,6 +242,23 @@ func (tx *TxObjectStorage) Rollback() error {

type ReferenceStorage map[plumbing.ReferenceName]*plumbing.Reference

+func (r ReferenceStorage) SetLog(ref *plumbing.Reference) error {
+	if ref != nil {
+
+		d := r[ref.Name()]
+		if d == nil {
+			d = ref
+		} else {
+			v := string(ref.Target()) + "\n" + string(d.Target())
+			d = plumbing.NewLogReference(ref.Name(), plumbing.ReferenceName(v))
+		}
+
+		r[ref.Name()] = d
+	}
+
+	return nil
     +}
+
func (r ReferenceStorage) SetReference(ref *plumbing.Reference) error {
if ref != nil {
r[ref.Name()] = ref
@@ -273,6 +291,18 @@ func (r ReferenceStorage) Reference(n plumbing.ReferenceName) (*plumbing.Referen
return ref, nil
}

+func (r ReferenceStorage) RefLog(n plumbing.ReferenceName, resolved bool) (*plumbing.Reference, error) {
+	ref, ok := r[n]
+	if !ok {
+		return nil, os.ErrNotExist
+	}
+
+	if resolved {
+		ref = plumbing.NewSymbolicReference(n, "true")
+	}
+	return ref, nil
     +}
+
func (r ReferenceStorage) IterReferences() (storer.ReferenceIter, error) {
var refs []*plumbing.Reference
for _, ref := range r {

## 9 utils/ioutil/pipe_js.go

diff --git a/utils/ioutil/pipe_js.go b/utils/ioutil/pipe_js.go
index cf102e6e..5671c920 100644
--- a/utils/ioutil/pipe_js.go
+++ b/utils/ioutil/pipe_js.go
@@ -1,8 +1,9 @@
+//go:build js
// +build js

package ioutil

-import "github.com/acomagu/bufpipe"
+import "github.com/go-git/go-git/v5/bufpipe"

func Pipe() (PipeReader, PipeWriter) {
return bufpipe.New(nil)

```