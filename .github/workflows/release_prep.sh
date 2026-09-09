#!/usr/bin/env bash
# Builds the release archive and prints the release notes to stdout.
#
# Called by bazel-contrib/.github/.github/workflows/release_ruleset.yaml, which
# requires this exact path and reads the release notes from stdout.
set -o errexit -o nounset -o pipefail

TAG="$1"
VERSION="${TAG#v}"
REPO_NAME="xrootfs"
ARCHIVE="${REPO_NAME}-${TAG}.zip"

# These exclusions reproduce the ones thedoctor0/zip-release used before this
# script replaced it, so the archive keeps the same file list.
#
# `--symlinks` is what keeps this fast and correct. Without it `zip` follows
# every symlink and stores what it points at, and it walks the tree before it
# applies `-x`, so an exclusion pattern does not stop the walk. With the bazel
# convenience symlinks present that means descending the whole output base:
# measured here, the command had not finished after three minutes and had
# written nothing. With `--symlinks` the same command takes 0.04s.
#
# The exclusions still matter, so the symlinks are left out rather than stored
# as links. `*bazel-*` and not `bazel-*`, because `bazel-*` only matches the
# top level and `integration` is its own workspace with its own
# `integration/bazel-bin` and friends.
#
# `release_notes.txt` must be excluded too. The reusable workflow runs this
# script as `release_prep.sh TAG > release_notes.txt`, so the shell creates
# that file in the working directory before the script starts.
zip --quiet --symlinks --recurse-paths "${ARCHIVE}" . \
  -x '*.git*' '/*node_modules/*' '.editorconfig' \
     '*bazel-*' \
     'release_notes.txt' "${ARCHIVE}"

cat <<NOTES
## Using Bzlmod

\`\`\`starlark
bazel_dep(name = "xrootfs", version = "${VERSION}")
\`\`\`
NOTES
