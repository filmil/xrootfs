package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// pruneEmptyDirs removes every directory under rootfs that has nothing in it,
// and returns what it removed, relative to rootfs.
//
// A rootfs is consumed as a Bazel tree artifact, and Bazel's disk and remote
// caches do not keep an empty directory inside one. A rootfs restored from
// the cache therefore differs from the one that was built: every empty
// directory is missing, and a symlink that pointed at one is dangling there
// and nowhere else. Removing them at extraction makes the cached rootfs the
// same as the fresh one, so anything that goes wrong goes wrong every time.
//
// Removing a directory can leave its parent empty, so this repeats until a
// pass removes nothing. The root itself is never removed.
func pruneEmptyDirs(rootfs string) ([]string, error) {
	var all []string
	for {
		var pass []string
		err := filepath.Walk(rootfs, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() || path == rootfs {
				return nil
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return fmt.Errorf("readdir %q: %w", path, err)
			}
			if len(entries) == 0 {
				pass = append(pass, path)
			}
			return nil
		})
		if err != nil {
			return all, err
		}
		if len(pass) == 0 {
			sort.Strings(all)
			return all, nil
		}
		for _, p := range pass {
			if err := os.Remove(p); err != nil {
				return all, fmt.Errorf("remove %q: %w", p, err)
			}
			all = append(all, rel(rootfs, p))
		}
	}
}

func rel(base, path string) string {
	r, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return r
}
