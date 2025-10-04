// This program extracts an OCI TAR file into a rootfs.
package main

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type DockerManifest []struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type OCIIndex struct {
	Manifests []struct {
		Digest string `json:"digest"`
	} `json:"manifests"`
}

type OCIManifest struct {
	Layers []struct {
		Digest string `json:"digest"`
	} `json:"layers"`
}

// Extract tarball to dest, preserving metadata
func extractTar(tarPath, dest string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("while opening tar file: %q: %w", tarPath, err)
	}
	defer f.Close()

	var r io.Reader = f

	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return fmt.Errorf("while creating target: %q: %q: %w", target, tarPath, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("while creating dir: %q: %q: %w", target, tarPath, err)
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("while opening dir: %q: %q: %w", target, tarPath, err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return fmt.Errorf("while copying: %q: %q", tarPath, err)
			}
			out.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			os.Symlink(hdr.Linkname, target)
		case tar.TypeLink: // hard link
			os.Link(filepath.Join(dest, hdr.Linkname), target)
		case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			// requires root
			mode := uint32(hdr.Mode & 07777)
			dev := int(hdr.Devmajor)<<8 | int(hdr.Devminor)
			if err := syscall.Mknod(target, mode, dev); err != nil {
				// ignore if not root
			}
		}

		// set ownership (requires root)
		_ = os.Lchown(target, hdr.Uid, hdr.Gid)
		// set times
		atime := hdr.AccessTime
		mtime := hdr.ModTime
		if atime.IsZero() {
			atime = time.Now()
		}
		if mtime.IsZero() {
			mtime = time.Now()
		}
		_ = os.Chtimes(target, atime, mtime)
	}
	return nil
}

// Apply whiteouts
func applyWhiteouts(layerTmp, rootfs string) error {
	return filepath.Walk(layerTmp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error while walking: %q: %w", path, err)
		}
		name := info.Name()
		rel, _ := filepath.Rel(layerTmp, path)

		if name == ".wh..wh..opq" {
			targetDir := filepath.Join(rootfs, filepath.Dir(rel))
			entries, _ := os.ReadDir(targetDir)
			for _, e := range entries {
				os.RemoveAll(filepath.Join(targetDir, e.Name()))
			}
			os.Remove(path)
			return nil
		}

		if strings.HasPrefix(name, ".wh.") {
			orig := strings.TrimPrefix(name, ".wh.")
			target := filepath.Join(rootfs, filepath.Dir(rel), orig)
			os.RemoveAll(target)
			os.Remove(path)
		}
		return nil
	})
}

// Copy layerTmp into rootfs, preserving metadata
func copyLayer(layerTmp, rootfs string) error {
	return filepath.Walk(layerTmp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(layerTmp, path)
		if rel == "." {
			return nil
		}
		dst := filepath.Join(rootfs, rel)

		// handle dirs
		if info.IsDir() {
			if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
				return err
			}
			if st, ok := info.Sys().(*syscall.Stat_t); ok {
				_ = os.Lchown(dst, int(st.Uid), int(st.Gid))
			}
			return nil
		}

		// symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(path)
			os.RemoveAll(dst)
			return os.Symlink(target, dst)
		}

		// regular files
		if info.Mode().IsRegular() {
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				return err
			}
			in, err := os.Open(path)
			if err != nil {
				return err
			}
			defer in.Close()
			out, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, info.Mode().Perm())
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, in); err != nil {
				out.Close()
				return err
			}
			out.Close()
			if st, ok := info.Sys().(*syscall.Stat_t); ok {
				_ = os.Lchown(dst, int(st.Uid), int(st.Gid))
				_ = os.Chtimes(dst, time.Now(), time.Unix(int64(st.Mtim.Sec), int64(st.Mtim.Nsec)))
			}
		}
		return nil
	})
}

func processLayers(layers []string, baseDir, rootfs string) error {
	for _, layer := range layers {
		layerPath := filepath.Join(baseDir, layer)
		layerTmp, _ := os.MkdirTemp("", "layer")
		defer os.RemoveAll(layerTmp)

		if err := extractTar(layerPath, layerTmp); err != nil {
			return fmt.Errorf("while extracting layer: %q: %w", layerPath, err)
		}
		if err := applyWhiteouts(layerTmp, rootfs); err != nil {
			return fmt.Errorf("while applying whiteouts: %q: %w", layerPath, err)
		}
		if err := copyLayer(layerTmp, rootfs); err != nil {
			return fmt.Errorf("while copying layer: %q: %w", layerPath, err)
		}
	}
	return nil
}

func run(imageTar, rootfs string) error {
	tmp, err := os.MkdirTemp("", "img")
	if err != nil {
		return fmt.Errorf("while creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmp)

	if err := extractTar(imageTar, tmp); err != nil {
		return fmt.Errorf("extractTar: %w", err)
	}

	// Docker save?
	if data, err := os.ReadFile(filepath.Join(tmp, "manifest.json")); err == nil {
		var manifest DockerManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
		layers := manifest[0].Layers
		if err := os.MkdirAll(rootfs, 0755); err != nil {
			return fmt.Errorf("MkdirAll: %w", err)
		}
		processLayers(layers, tmp, rootfs)
		return nil
	}

	// OCI archive?
	if _, err := os.Stat(filepath.Join(tmp, "oci-layout")); err == nil {
		idxData, err := os.ReadFile(filepath.Join(tmp, "index.json"))
		if err != nil {
			return fmt.Errorf("while reading index: %w", err)
		}
		var idx OCIIndex
		if err := json.Unmarshal(idxData, &idx); err != nil {
			return fmt.Errorf("while unmarshalling index: %w", err)
		}
		if len(idx.Manifests) == 0 {
			return fmt.Errorf("no manifests in index.json (??)")
		}
		digest := idx.Manifests[0].Digest
		sha := strings.TrimPrefix(digest, "sha256:")
		manifestPath := filepath.Join(tmp, "blobs", "sha256", sha)
		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("while reading manifest: %q: %w", manifestPath, err)
		}
		var mf OCIManifest
		if err := json.Unmarshal(manifestData, &mf); err != nil {
			return fmt.Errorf("while unmarshalling manifest: %q: %w", manifestPath, err)
		}
		var layers []string
		for _, l := range mf.Layers {
			sha := strings.TrimPrefix(l.Digest, "sha256:")
			layers = append(layers, filepath.Join("blobs", "sha256", sha))
		}
		if err := os.MkdirAll(rootfs, 0755); err != nil {
			return fmt.Errorf("while creating dir for: %q: %w", rootfs, err)
		}
		processLayers(layers, tmp, rootfs)
		return nil
	}

	return fmt.Errorf("unrecognized archive format (not Docker save or OCI archive)")
}

func main() {
	prgname := os.Args[0]
	log.SetPrefix(fmt.Sprintf("%v:\n\t", prgname))
	var (
		imageTar, rootfs string
	)
	flag.StringVar(&imageTar, "image-tar", "", "The TAR archive of an OCI image file")
	flag.StringVar(&rootfs, "rootfs-dir", "", "The name of the directory to put the extracted rootfs in")
	flag.Parse()

	if imageTar == "" {
		log.Printf("flag --image-tar=... is mandatory")
		os.Exit(1)
	}
	if rootfs == "" {
		log.Printf("flag --rootfs-dir=... is mandatory")
		os.Exit(1)
	}

	if err := run(imageTar, rootfs); err != nil {
		log.Printf("error while processing %q into %q: %v", imageTar, rootfs, err)
		os.Exit(1)
	}
}
