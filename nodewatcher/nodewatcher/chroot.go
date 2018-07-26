package nodewatcher

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Chroot changes the root of `path` using `newRoot`.
// Chroot returns an error if :
// `newRoot` and `path` are not absolute paths.
// `newRoot` is not in the absolute path of `path`.
func Chroot(path, newRoot string) (string, error) {
	path = filepath.Clean(path)
	newRoot = filepath.Clean(newRoot)
	if !filepath.IsAbs(newRoot) {
		return "", fmt.Errorf("`%s` is not an absolute path", newRoot)
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("`%s` is not an absolute path", path)
	}
	if !strings.HasPrefix(path, newRoot) {
		return "", fmt.Errorf("`%s` is not part of the path `%s`", newRoot, path)
	}
	if newRoot == "/" {
		return path, nil
	}
	return strings.TrimPrefix(strings.TrimPrefix(path, filepath.Dir(newRoot)), string(filepath.Separator)), nil
}
