package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestChroot(t *testing.T) {
	_, err := Chroot("a", "/b")
	if err == nil || err.Error() != "`a` is not an absolute path" {
		t.Fatal("Chroot should return an error when `path` is not an absolute path")
	}
	_, err = Chroot("/a", "b")
	if err == nil || err.Error() != "`b` is not an absolute path" {
		t.Fatal("Chroot should return an error when `newRoot` is not an absolute path")
	}
	_, err = Chroot("/c/a", "/a")
	if err == nil || err.Error() != "`/a` is not part of the path `/c/a`" {
		t.Fatal("Chroot should return an error when `newRoot` is not part of `path`")
	}
	newPath, err := Chroot("/c/a", "/c/a")
	if err != nil {
		t.Fatal(err)
	}
	if newPath != "a" {
		t.Fatalf("newPath should be a, is `%s` instead", newPath)
	}
	newPath, err = Chroot("/a/b/c/d", "/a/b")
	if err != nil {
		t.Fatal(err)
	}
	if newPath != "b/c/d" {
		t.Fatalf("newPath should be b/c/d, is `%s` instead", newPath)
	}
	newPath, err = Chroot("/a/b", "/a")
	if err != nil {
		t.Fatal(err)
	}
	if newPath != "a/b" {
		t.Fatalf("newPath should be a/b, is `%s` instead", newPath)
	}
	newPath, err = Chroot("/a/b", "/")
	if err != nil {
		t.Fatal(err)
	}
	if newPath != "/a/b" {
		t.Fatalf("newPath should be /a/b, is `%s` instead", newPath)
	}
	newPath, err = Chroot("/a/b/c/d", "/a/b/c/../.")
	if err != nil {
		t.Fatal(err)
	}
	if newPath != "b/c/d" {
		t.Fatalf("newPath should be b/c/d, is `%s` instead", newPath)
	}
}
func TestListFiles(t *testing.T) {
	files := make(map[string]bool)
	tmpDir, err := ioutil.TempDir("", "nodewatcher")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	files[path.Base(tmpDir)] = true
	// Create a file in tmpDir
	tmpFile, err := ioutil.TempFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	newPath, err := Chroot(tmpFile.Name(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	files[newPath] = true
	// Create a subdirectory in tmpDir
	subTmpDir, err := ioutil.TempDir(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	newPath, err = Chroot(subTmpDir, tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	files[newPath] = true
	// Create a file in that subdirectory
	tmpFile, err = ioutil.TempFile(subTmpDir, "")
	if err != nil {
		t.Fatal(err)
	}
	newPath, err = Chroot(tmpFile.Name(), tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	files[newPath] = true
	pathCh := make(chan string)
	go ListFiles(tmpDir, pathCh)
	for path := range pathCh {
		if _, ok := files[path]; !ok {
			t.Fatalf("`%s` should be in the list of files : `%v`", path, files)
		}
	}
}
