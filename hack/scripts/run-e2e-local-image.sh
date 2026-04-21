#!/usr/bin/env bash

# Copyright The Guard Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

# Usage ------------------------------------------------------------------------

# usage prints a short help message describing environment variables and common
# invocation patterns.
usage() {
    cat <<'EOF'
Usage:
  hack/scripts/run-e2e-local-image.sh [ginkgo/test args...]

Build a local Guard image, load it into a Kind cluster, and run the E2E suite
against that image.

Environment variables:
  GUARD_IMAGE          Image reference to build and test. Unqualified images are
                       normalized to localhost/... (default: guard-e2e/guard:local)
  CONTAINER_BIN        Container CLI to use for build/save (auto-detects podman,
                       docker, or docker.exe)
  KIND_BIN             Kind CLI to use (auto-detects kind or kind.exe)
  KUBE_CONTEXT         Kubernetes context for kubectl and E2E
  KIND_CLUSTER_NAME    Kind cluster name. Defaults from kind-* kube context or "kind"
  KIND_LOAD_STRATEGY   One of: auto, docker-image, image-archive (default: auto)
  GINKGO_FOCUS         Focus regex passed to Ginkgo
  KUBECONFIG           Optional kubeconfig path for the E2E binary
  ARCHIVE_PATH         Optional archive path when KIND_LOAD_STRATEGY=image-archive
  BINARY_PATH          Output path for the built Guard Linux binary
  DOCKERFILE_PATH      Output path for the rendered container Dockerfile
  BASEIMAGE_PROD       Base image used when rendering the production Dockerfile.
                       Defaults to the repo Makefile's BASEIMAGE_PROD value.

Examples:
  hack/scripts/run-e2e-local-image.sh
  GINKGO_FOCUS='Set up guard for github should be successful' \
    hack/scripts/run-e2e-local-image.sh
  CONTAINER_BIN=docker KIND_LOAD_STRATEGY=docker-image \
    hack/scripts/run-e2e-local-image.sh
  GUARD_IMAGE=localhost/guard-e2e/guard:dev \
    hack/scripts/run-e2e-local-image.sh
  CONTAINER_BIN=podman KIND_BIN=kind.exe \
    KUBE_CONTEXT=kind-kind-cluster KIND_CLUSTER_NAME=kind-cluster \
    KIND_LOAD_STRATEGY=image-archive \
    hack/scripts/run-e2e-local-image.sh
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    usage
    exit 0
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

# Configuration ----------------------------------------------------------------

IMAGE_NAME="${GUARD_IMAGE:-guard-e2e/guard:local}"
BINARY_PATH="${BINARY_PATH:-bin/guard-linux-amd64}"
DOCKERFILE_PATH="${DOCKERFILE_PATH:-bin/.dockerfile-PROD-linux_amd64}"
GINKGO_FOCUS="${GINKGO_FOCUS:-}"
KIND_LOAD_STRATEGY="${KIND_LOAD_STRATEGY:-auto}"
TEMP_ARCHIVE=0
ARCHIVE_PATH="${ARCHIVE_PATH:-}"

if [[ -z "$ARCHIVE_PATH" ]]; then
    ARCHIVE_PATH="$(mktemp /tmp/guard-e2e-image-XXXXXX.tar)"
    TEMP_ARCHIVE=1
fi

cleanup() {
    if [[ "$TEMP_ARCHIVE" == "1" && -f "$ARCHIVE_PATH" ]]; then
        rm -f "$ARCHIVE_PATH"
    fi
}
trap cleanup EXIT

# Helpers ----------------------------------------------------------------------

# require_command exits early when a required executable is not available.
require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "Required command not found: $1" >&2
        exit 1
    fi
}

# resolve_command returns the first command from the candidate list that exists
# in PATH.
resolve_command() {
    local cmd

    for cmd in "$@"; do
        if [[ -n "$cmd" ]] && command -v "$cmd" >/dev/null 2>&1; then
            printf '%s' "$cmd"
            return 0
        fi
    done

    return 1
}

# default_kind_cluster_name derives a Kind cluster name from a kind-* kube
# context, or falls back to Kind's default cluster name.
default_kind_cluster_name() {
    local kube_context_name="$1"

    if [[ "$kube_context_name" == kind-* ]]; then
        printf '%s' "${kube_context_name#kind-}"
        return
    fi

    printf 'kind'
}

# resolve_make_optional_var reads a simple `VAR ?= value` assignment from the
# repo Makefile so the script can follow existing build defaults.
resolve_make_optional_var() {
    local var_name="$1"

    # Match only optional assignments of the requested variable, then strip the
    # variable name and operator before printing the remainder of the line.
    awk -v var_name="$var_name" '
    $1 == var_name && $2 == "?=" {
      $1 = ""
      $2 = ""
      sub(/^[[:space:]]+/, "")
      print
      exit
    }
  ' Makefile
}

# normalize_image_name rewrites unqualified images to localhost/... so local
# kind nodes can resolve the same reference we build and load.
normalize_image_name() {
    local image="$1"
    local first_component="${image%%/*}"

    if [[ "$image" != */* ]]; then
        printf 'localhost/library/%s' "$image"
        return
    fi

    if [[ "$first_component" == "localhost" || "$first_component" == *.* || "$first_component" == *:* ]]; then
        printf '%s' "$image"
        return
    fi

    printf 'localhost/%s' "$image"
}

# resolve_kind_load_strategy picks a supported Kind image-loading mode. Archive
# loading is the safest cross-platform default, while docker-image works for the
# common local Docker + Linux kind setup.
resolve_kind_load_strategy() {
    local strategy="$1"
    local container_bin_name kind_bin_name

    case "$strategy" in
        docker-image | image-archive)
            printf '%s' "$strategy"
            return
            ;;
        auto)
            container_bin_name="${CONTAINER_BIN##*/}"
            kind_bin_name="${KIND_BIN##*/}"

            if [[ "$container_bin_name" == docker || "$container_bin_name" == docker.exe ]] &&
                [[ "$kind_bin_name" != *.exe ]]; then
                printf 'docker-image'
                return
            fi

            printf 'image-archive'
            return
            ;;
        *)
            echo "Unsupported KIND_LOAD_STRATEGY: $strategy" >&2
            exit 1
            ;;
    esac
}

# kind_archive_path converts the archive path for Windows kind.exe, while Linux
# kind can consume the native filesystem path directly.
kind_archive_path() {
    if [[ "${KIND_BIN##*/}" == *.exe ]]; then
        require_command wslpath
        wslpath -w "$ARCHIVE_PATH"
        return
    fi

    printf '%s' "$ARCHIVE_PATH"
}

# Tool detection ----------------------------------------------------------------

require_command go
require_command kubectl
require_command awk

CONTAINER_BIN="${CONTAINER_BIN:-}"
if [[ -z "$CONTAINER_BIN" ]]; then
    CONTAINER_BIN="$(resolve_command podman docker docker.exe || true)"
fi
if [[ -z "$CONTAINER_BIN" ]]; then
    echo "Unable to locate a container CLI. Set CONTAINER_BIN or install podman/docker." >&2
    exit 1
fi

KIND_BIN="${KIND_BIN:-}"
if [[ -z "$KIND_BIN" ]]; then
    KIND_BIN="$(resolve_command kind kind.exe || true)"
fi
if [[ -z "$KIND_BIN" ]]; then
    echo "Unable to locate kind. Set KIND_BIN or install kind." >&2
    exit 1
fi

require_command "$CONTAINER_BIN"
require_command "$KIND_BIN"

BASEIMAGE_PROD="${BASEIMAGE_PROD:-$(resolve_make_optional_var BASEIMAGE_PROD)}"
if [[ -z "$BASEIMAGE_PROD" ]]; then
    echo "Unable to determine BASEIMAGE_PROD from the environment or Makefile." >&2
    exit 1
fi

# Prefer explicit overrides, otherwise inherit the current kubectl context and
# derive a Kind cluster name that matches it.
CURRENT_KUBE_CONTEXT="$(kubectl config current-context 2>/dev/null || true)"
KUBE_CONTEXT_NAME="${KUBE_CONTEXT:-$CURRENT_KUBE_CONTEXT}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-$(default_kind_cluster_name "$KUBE_CONTEXT_NAME")}"
IMAGE_NAME="$(normalize_image_name "$IMAGE_NAME")"
KIND_LOAD_STRATEGY="$(resolve_kind_load_strategy "$KIND_LOAD_STRATEGY")"

if [[ -n "$KUBE_CONTEXT_NAME" ]]; then
    kubectl --context "$KUBE_CONTEXT_NAME" cluster-info >/dev/null
else
    kubectl cluster-info >/dev/null
fi

# Build metadata ----------------------------------------------------------------

mkdir -p bin

git_branch="$(git rev-parse --abbrev-ref HEAD)"
git_tag="$(git describe --exact-match --abbrev=0 2>/dev/null || true)"
commit_hash="$(git rev-parse --verify HEAD)"
commit_timestamp="$(date --date="@$(git show -s --format=%ct)" --utc +%FT%T)"
version="$(git describe --tags --always --dirty)"
version_strategy="commit_hash"
go_version="$(go version | cut -d ' ' -f 3)"

if [[ -n "$git_tag" ]]; then
    version="$git_tag"
    version_strategy="tag"
fi

ldflags=(
    "-X main.Version=$version"
    "-X main.VersionStrategy=$version_strategy"
    "-X main.GitTag=$git_tag"
    "-X main.GitBranch=$git_branch"
    "-X main.CommitHash=$commit_hash"
    "-X main.CommitTimestamp=$commit_timestamp"
    "-X main.GoVersion=$go_version"
    "-X main.Compiler=$(go env CC)"
    "-X main.Platform=linux/amd64"
)
# Join the ldflags array into a single space-delimited string for go build.
printf -v GO_LDFLAGS '%s ' "${ldflags[@]}"

# Build binary and image --------------------------------------------------------

echo "Building Guard binary at $BINARY_PATH"
CGO_ENABLED=0 \
    GO111MODULE=on \
    GOFLAGS='-mod=vendor' \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o "$BINARY_PATH" -ldflags "$GO_LDFLAGS" .

echo "Rendering container Dockerfile at $DOCKERFILE_PATH"
sed \
    -e 's|{ARG_BIN}|guard|g' \
    -e 's|{ARG_ARCH}|amd64|g' \
    -e 's|{ARG_OS}|linux|g' \
    -e "s|{ARG_FROM}|$BASEIMAGE_PROD|g" \
    Dockerfile.in >"$DOCKERFILE_PATH"

echo "Building local image $IMAGE_NAME with $CONTAINER_BIN"
"$CONTAINER_BIN" build -t "$IMAGE_NAME" -f "$DOCKERFILE_PATH" .

# Load image into kind ----------------------------------------------------------

case "$KIND_LOAD_STRATEGY" in
    docker-image)
        echo "Loading image into kind cluster $KIND_CLUSTER_NAME via docker-image"
        "$KIND_BIN" load docker-image "$IMAGE_NAME" --name "$KIND_CLUSTER_NAME"
        ;;
    image-archive)
        echo "Saving image archive to $ARCHIVE_PATH"
        "$CONTAINER_BIN" save -o "$ARCHIVE_PATH" "$IMAGE_NAME"

        # kind.exe expects a Windows path, while Linux kind expects the original
        # POSIX path. kind_archive_path handles that translation.
        echo "Loading image archive into kind cluster $KIND_CLUSTER_NAME"
        "$KIND_BIN" load image-archive "$(kind_archive_path)" --name "$KIND_CLUSTER_NAME"
        ;;
esac

# Run tests --------------------------------------------------------------------

ginkgo_cmd=(go run github.com/onsi/ginkgo/v2/ginkgo -r --v --show-node-events --trace)
if [[ -n "$GINKGO_FOCUS" ]]; then
    ginkgo_cmd+=("--focus=$GINKGO_FOCUS")
fi

ginkgo_cmd+=(./test/e2e --)
if [[ -n "$KUBE_CONTEXT_NAME" ]]; then
    ginkgo_cmd+=(--kube-context "$KUBE_CONTEXT_NAME")
fi

ginkgo_cmd+=(--guard-image "$IMAGE_NAME")
if [[ -n "${KUBECONFIG:-}" ]]; then
    ginkgo_cmd+=(--kubeconfig "$KUBECONFIG")
fi
if [[ "$#" -gt 0 ]]; then
    ginkgo_cmd+=("$@")
fi

echo "Running e2e tests:"
printf '  %q' "${ginkgo_cmd[@]}"
printf '\n'
"${ginkgo_cmd[@]}"
