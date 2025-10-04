# xrootfs

[![Test](https://github.com/filmil/xrootfs/actions/workflows/test.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/test.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml)
[![Publish to my Bazel registry](https://github.com/filmil/xrootfs/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish.yml)
[![Tag and Release](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml)

A program to extract rootfs into a directory, from an OCI TAR file.

I can't believe there doesn't seem to be such a thing already out there. There
are programs which rely on docker or pod, but apparently no standalone binaries.

Well, here goes change.

