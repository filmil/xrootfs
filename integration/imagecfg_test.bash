#!/usr/bin/env bash
# Renders the bundled template with imagecfg and checks the result.
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

imagecfg="$(rlocation "$1")"
template="$(rlocation integration/image.yml.tpl)"
out="${TEST_TMPDIR}/image.yml"

"${imagecfg}" \
  --template="${template}" \
  --output="${out}" \
  --arch=amd64 \
  --package=cpio \
  --package=bash \
  --source=https://snapshot.ubuntu.com/=noble,main,universe

echo "--- rendered:"
cat "${out}"

expected="${TEST_TMPDIR}/expected.yml"
cat >"${expected}" <<'EOF'
version: 1
sources:
  - channel: noble main universe
    url: https://snapshot.ubuntu.com/
archs:
  - amd64
packages:
  - cpio
  - bash
EOF

# -B ignores the trailing blank line the template emits after the last
# package. The content is what this asserts, not the whitespace at the end.
diff -u -B "${expected}" "${out}"
echo "imagecfg test passed!"
