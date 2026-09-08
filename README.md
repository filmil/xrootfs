# xrootfs

[![Test](https://github.com/filmil/xrootfs/actions/workflows/test.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/test.yml)
[![Publish on Bazel Central Registry](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish-bcr.yml)
[![Publish to my Bazel registry](https://github.com/filmil/xrootfs/actions/workflows/publish.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/publish.yml)
[![Tag and Release](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml/badge.svg)](https://github.com/filmil/xrootfs/actions/workflows/tag-and-release.yml)

Turn an OCI or Docker image archive into a plain directory tree, without a
container runtime.

## Why this exists

Unpacking a container image is normally done by asking a runtime to do it.
`docker export`, `podman unshare`, `umoci` and `skopeo` all work, and all of
them want a daemon, a privileged helper, or a large dependency on the build
machine.

That is a poor fit for a hermetic build. A Bazel action that shells out to
Docker depends on the state of the machine it runs on, cannot run in a sandbox
that has no daemon socket, and produces a different answer on a machine where
the daemon is configured differently.

`xrootfs` is a single static Go binary that reads the image archive itself and
writes the files out. It needs no daemon, no root, and no network. A build can
depend on it the same way it depends on any other tool.

## The two programs

| Program | What it does |
| --- | --- |
| `xrootfs` | Extracts a rootfs directory from an OCI or Docker image TAR |
| `imagecfg` | Renders an image configuration file from a list of packages, architectures and package sources |

## Installing

### As a Bazel module

The module is published to a personal registry, so add that registry to
`.bazelrc` first. Adding a registry replaces the default, so name the Bazel
Central Registry as well or the rest of your dependencies stop resolving:

```
common --registry=https://bcr.bazel.build
common --registry=https://raw.githubusercontent.com/filmil/bazel-registry/main
```

Then depend on it from `MODULE.bazel`:

```starlark
bazel_dep(name = "xrootfs", version = "0.3.0")  # Choose the version here.
```

Both binaries are public targets, so a rule can use either as a tool:

```starlark
run_binary(
    name = "rootfs",
    tool = "@xrootfs//cmd/xrootfs",
    # ...
)
```

### As a prebuilt binary

Every release attaches `linux-386` and `linux-amd64` zips for both programs.
See the [releases page](https://github.com/filmil/xrootfs/releases).

### From source

```console
git clone https://github.com/filmil/xrootfs.git
cd xrootfs
bazel build //cmd/xrootfs //cmd/imagecfg
bazel test //...
```

## Using `xrootfs`

Give it an image archive and a directory to write:

```console
xrootfs --image-tar=image.tar --rootfs-dir=./rootfs
```

It reads both the Docker manifest format and the OCI index format, applies
every layer in order, and preserves permissions, symlinks, hard links and
device nodes. Device nodes need root; everything else does not.

### Flags

```
  -image-tar string
        The TAR archive of an OCI image file (required)
  -rootfs-dir string
        The name of the directory to put the extracted rootfs in (required)
  -fix-links
        Whether to fix dangling links or not (default true)
  -marker string
        The name of a marker file to create in rootfs - skipped if empty
  -rm value
        One entry for each file (relative to rootfs) to delete. Repeatable
  -prune-dangling-links
        Whether to remove symlinks whose target does not exist in the rootfs
        (default true)
  -prune-empty-dirs
        Whether to remove empty directories, which a Bazel cache would drop
        anyway (default true)
```

### What the defaults are for

The three defaults that rewrite the tree exist because the output is meant to
be consumed as a Bazel tree artifact.

`--fix-links` rewrites absolute symlinks to be relative to the rootfs
directory. An image that ships `/usr/bin/sh -> /bin/sh` means `/bin/sh` inside
the image. Extracted under `./rootfs`, that link would point at the build
machine's own `/bin/sh`, so it is rewritten to reach `./rootfs/bin/sh`
instead.

`--prune-empty-dirs` removes directories with nothing in them. Bazel's disk
and remote caches do not keep an empty directory inside a tree artifact, so a
rootfs restored from cache would otherwise differ from the one just built.
Removing them at extraction time makes both the same, which means anything
that goes wrong goes wrong every time rather than only on a cache miss.

`--prune-dangling-links` removes symlinks whose target is not in the rootfs.
These are usually links into a directory that pruning has just removed, or
into a path the image expects the runtime to mount.

`--marker` writes an empty file at a known path. A Bazel rule can declare that
file as its output when the rest of the tree is a directory whose contents are
not known ahead of time.

## Using `imagecfg`

`imagecfg` fills in a template from repeated command line flags. It exists so
that the package list for an image can live in a `BUILD.bazel` file as
ordinary Starlark lists, rather than in a hand maintained YAML file.

```console
imagecfg \
  --template=image.yml.tpl \
  --output=image.yml \
  --arch=amd64 \
  --package=cpio \
  --package=bash \
  --source=https://snapshot.ubuntu.com/=noble,main,universe
```

With the bundled `image.yml.tpl`, that writes:

```yaml
version: 1
sources:
  - channel: noble main universe
    url: https://snapshot.ubuntu.com/
archs:
  - amd64
packages:
  - cpio
  - bash
```

### Flags

```
  -template string
        The template file to render (required)
  -output string
        The file to write the rendered result to (required)
  -package value
        A package to include, such as `cpio`. Repeatable
  -arch value
        An arch to include, such as `amd64`. Repeatable
  -source value
        Map from URL to a comma separated list of channels, such as
        --source=https://snapshot.ubuntu.com/=noble,main,universe. Repeatable
```

The template is an ordinary Go template. It is given `.Archs`, `.Packages`
and `.Sources`, where each source has `.URL` and `.Channels`.

## Releases and provenance

Releases are cut by the `Tag and Release` workflow in
`.github/workflows/tag-and-release.yml`, which runs monthly and on demand.

The source archive is built and attested by the reusable release workflow from
`bazel-contrib/.github`, so each release also carries
`xrootfs-<tag>.zip.intoto.jsonl`, a SLSA build provenance attestation. Verify
a downloaded archive with:

```console
gh attestation verify xrootfs-<tag>.zip --repo filmil/xrootfs
```

## Related work

Other tools cover overlapping ground, and are a better fit when a runtime is
available anyway:

* [umoci](https://github.com/opencontainers/umoci), manipulates OCI images.
* [skopeo](https://github.com/containers/skopeo), moves images between
  registries and formats.
* [podman](https://podman.io/), a daemonless container engine.
* [buildah](https://buildah.io/), builds OCI images.

## License

Apache 2.0. See [LICENSE](LICENSE).
