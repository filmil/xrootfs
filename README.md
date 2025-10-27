# xrootfs

[![Test](https://github.com/filmil/xrootfs/actions/workflows/test.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/test.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml)
[![Publish to my Bazel registry](https://github.com/filmil/xrootfs/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish.yml)
[![Tag and Release](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml)

`xrootfs` is a collection of standalone programs to extract a rootfs from an
OCI or Docker TAR archive and apply various bits of configuration to it.

There are programs which rely on docker or podman to extract OCI TAR files, but
apparently no standalone binaries that can do this on their own. This program
aims to fill that gap.

