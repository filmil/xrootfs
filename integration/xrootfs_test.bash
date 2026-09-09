#!/usr/bin/env bash
# Extracts a rootfs from a Docker format archive that this test builds, and
# checks what xrootfs did with the layers, the symlinks and the empty dirs.
set -o errexit -o nounset -o pipefail

# --- runfiles bootstrap, see @bazel_tools//tools/bash/runfiles ---
f=bazel_tools/tools/bash/runfiles/runfiles.bash
# shellcheck disable=SC1090,SC1091
source "${RUNFILES_DIR:-/dev/null}/$f" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "${RUNFILES_MANIFEST_FILE:-/dev/null}" | cut -f2- -d' ')" 2>/dev/null || \
  source "$0.runfiles/$f" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "$0.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
  source "$(grep -sm1 "^$f " "$0.exe.runfiles_manifest" | cut -f2- -d' ')" 2>/dev/null || \
  { echo>&2 "ERROR: cannot find $f"; exit 1; }; f=; set -e
# --- end runfiles bootstrap ---

xrootfs="$(rlocation "$1")"
work="${TEST_TMPDIR}/work"
mkdir -p "${work}"

# Build one image layer. It holds a regular file, an absolute symlink of the
# kind an image ships (/usr/bin/sh -> /bin/sh), a symlink to nothing, and a
# directory with nothing in it.
layer="${work}/layer"
mkdir -p "${layer}/bin" "${layer}/usr/bin" "${layer}/var/empty"
echo "hello from the image" > "${layer}/bin/sh"
ln -s /bin/sh "${layer}/usr/bin/sh"
ln -s /nowhere "${layer}/usr/bin/dangling"
tar --create --file "${work}/layer.tar" --directory "${layer}" .

# Wrap it as a Docker format archive: layer tarballs plus a manifest.json.
cat > "${work}/manifest.json" <<'EOF'
[{"Config":"config.json","RepoTags":["test:latest"],"Layers":["layer.tar"]}]
EOF
echo '{}' > "${work}/config.json"
tar --create --file "${work}/image.tar" --directory "${work}" \
    manifest.json config.json layer.tar

rootfs="${TEST_TMPDIR}/rootfs"
"${xrootfs}" --image-tar="${work}/image.tar" --rootfs-dir="${rootfs}" \
             --marker=.extracted

echo "--- extracted tree:"
( cd "${rootfs}" && find . | sort )

fail() { echo "FAIL: $*" >&2; exit 1; }

# The layer's regular file is there, with its contents.
[[ -f "${rootfs}/bin/sh" ]] || fail "bin/sh missing"
grep -q "hello from the image" "${rootfs}/bin/sh" || fail "bin/sh has wrong contents"

# --marker wrote its file.
[[ -f "${rootfs}/.extracted" ]] || fail ".extracted marker missing"

# --fix-links rewrote the absolute symlink to stay inside the rootfs, so it
# resolves to the layer's own file rather than the host's /bin/sh.
[[ -L "${rootfs}/usr/bin/sh" ]] || fail "usr/bin/sh is not a symlink"
target="$(readlink "${rootfs}/usr/bin/sh")"
[[ "${target}" != /* ]] || fail "usr/bin/sh still absolute: ${target}"
grep -q "hello from the image" "${rootfs}/usr/bin/sh" \
  || fail "usr/bin/sh does not resolve inside the rootfs, points at ${target}"

# --prune-dangling-links removed the symlink that pointed at nothing.
[[ ! -e "${rootfs}/usr/bin/dangling" && ! -L "${rootfs}/usr/bin/dangling" ]] \
  || fail "dangling symlink was not pruned"

# --prune-empty-dirs removed the directory with nothing in it.
[[ ! -d "${rootfs}/var/empty" ]] || fail "empty directory was not pruned"

echo "xrootfs test passed!"
