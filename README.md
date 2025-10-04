# xrootfs

[![Test](https://github.com/filmil/xrootfs/actions/workflows/test.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/test.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml)
[![Publish to my Bazel registry](https://github.com/filmil/xrootfs/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish.yml)
[![Tag and Release](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml)

`xrootfs` is a standalone program to extract a rootfs from an OCI or Docker TAR archive.

There are programs which rely on docker or podman to extract OCI TAR files, but apparently no standalone binaries that can do this on their own. This program aims to fill that gap.

## Flags

```
  -fix-links
    	Whether to fix dangling links or not (default true)
  -image-tar string
    	The TAR archive of an OCI image file
  -marker string
    	The name of a marker file to create in rootfs - skipped if empty
  -rm value
    	One entry for each file (relative to rootfs) to delete
  -rootfs-dir string
    	The name of the directory to put the extracted rootfs in
```

## References

There are other programs that do similar things. Here are a few of them:

* [umoci](https://github.com/opencontainers/umoci): A command line tool to manipulate OCI images.
* [skopeo](https://github.com/containers/skopeo): A command line utility that performs various operations on container images and image repositories.
* [podman](https://podman.io/): A daemonless container engine for developing, managing, and running OCI Containers on your Linux System.
* [buildah](https://buildah.io/): A tool that facilitates building OCI container images.