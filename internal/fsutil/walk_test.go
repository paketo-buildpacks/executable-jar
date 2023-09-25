package fsutil_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/paketo-buildpacks/executable-jar/v6/internal/fsutil"
)

func TestWalkAgainstStdlibWalk(t *testing.T) {
	type fileInfo struct {
		Path string
		Type fs.FileMode
	}
	createWalkFn := func(out *[]fileInfo) filepath.WalkFunc {
		return func(path string, fi fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			var typ fs.FileMode
			if fi != nil {
				typ = fi.Mode().Type()
			}
			*out = append(*out, fileInfo{
				Path: path,
				Type: typ,
			})
			return nil
		}
	}

	type args struct {
		root string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "basic",
			args: args{
				root: createTestDir(t),
			},
		},
		{
			name: "non-existent root",
			args: args{
				root: filepath.Join(t.TempDir(), "nonexistent"),
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			var ourFiles, theirsFiles []fileInfo
			errOurs := fsutil.Walk(tt.args.root, createWalkFn(&ourFiles))
			errTheirs := filepath.Walk(tt.args.root, createWalkFn(&theirsFiles))

			if (ourFiles == nil) != (theirsFiles == nil) {
				t.Errorf("errors does not match, actual error: %#v, expected error: %#v", errOurs, errTheirs)
			}

			// sort output of filepath.Walk in BFS order
			sort.SliceStable(theirsFiles, func(i, j int) bool {
				iLen := len(strings.Split(theirsFiles[i].Path, string(filepath.Separator)))
				jLen := len(strings.Split(theirsFiles[j].Path, string(filepath.Separator)))
				return iLen < jLen
			})

			if d := cmp.Diff(theirsFiles, ourFiles); d != "" {
				t.Error("output missmatch (-want, +got):", d)
			}
		})

	}
}

func TestWalkSearch(t *testing.T) {
	root := createTestDir(t)
	sentinelErr := errors.New("sentinel")
	var content string
	err := fsutil.Walk(root, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.Mode().Type() == 0 {
			_, name := filepath.Split(path)
			if name == "a.txt" {
				bs, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				content = string(bs)
				return sentinelErr
			}
		}
		return nil
	})
	if !errors.Is(err, sentinelErr) {
		t.Errorf("incorrect return value, expected sentinel error but got: %v", err)
	}
	if content != "a1/a2/a.txt" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestWalkSearchWithSkipDir(t *testing.T) {
	root := createTestDir(t)
	sentinelErr := errors.New("sentinel")
	var content string
	err := fsutil.Walk(root, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rp, err := filepath.Rel(root, path)
		if rp == "a1" { // skip the first dir containing a.txt
			return filepath.SkipDir
		}
		if fi.Mode().Type() == 0 {
			_, name := filepath.Split(path)
			if name == "a.txt" {
				bs, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				content = string(bs)
				return sentinelErr
			}
		}
		return nil
	})
	if !errors.Is(err, sentinelErr) {
		t.Errorf("incorrect return value, expected sentinel error but got: %v", err)
	}
	if content != "c1/a2/a3/a.txt" {
		t.Errorf("unexpected content: %q", content)
	}
}

func createTestDir(t *testing.T) string {
	root := t.TempDir()
	var err error

	var createChildrenDirs func(dir string, lvl int)
	createChildrenDirs = func(dir string, lvl int) {
		if lvl >= 4 {
			return
		}
		for _, n := range []string{"a", "b", "c"} {
			d := filepath.Join(dir, fmt.Sprintf("%s%d", n, lvl))
			err := os.Mkdir(d, 0755)
			if err != nil {
				t.Fatal(err)
			}
			createChildrenDirs(d, lvl+1)
		}
	}
	createChildrenDirs(root, 1)

	err = os.Chmod(filepath.Join(root, "b1", "a2"), 0000)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err = os.Chmod(filepath.Join(root, "b1", "a2"), 0755)
		if err != nil {
			t.Fatal(err)
		}
	})

	err = os.Symlink("a1", filepath.Join(root, "a.lnk"))
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(root, "a1", "a2", "a.txt"), []byte(`a1/a2/a.txt`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(root, "c1", "a2", "a3", "a.txt"), []byte(`c1/a2/a3/a.txt`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	return root
}
