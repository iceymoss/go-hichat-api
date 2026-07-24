#!/usr/bin/env bash

set -euo pipefail

readonly GOCTL_VERSION="1.8.2"
readonly PROTOC_VERSION="28.3"
readonly PROTOC_GEN_GO_VERSION="1.35.2"
readonly PROTOC_GEN_GO_GRPC_VERSION="1.5.1"

readonly -a CONTRACT_FILES=(
  apps/social/api/social.api
  apps/social/rpc/social.proto
  apps/im/api/im.api
  apps/im/rpc/im.proto
)
readonly -a GENERATED_FILES=(
  apps/social/api/internal/types/types.go
  apps/social/api/internal/handler/routes.go
  apps/social/rpc/social/social.pb.go
  apps/social/rpc/social/social_grpc.pb.go
  apps/social/rpc/socialclient/social.go
  apps/im/api/internal/types/types.go
  apps/im/api/internal/handler/routes.go
  apps/im/rpc/im/im.pb.go
  apps/im/rpc/im/im_grpc.pb.go
  apps/im/rpc/imclient/im.go
)

committed=false
case ${1:-} in
  "") ;;
  --committed) committed=true ;;
  *)
    printf 'usage: %s [--committed]\n' "$0" >&2
    exit 2
    ;;
esac

fail_version() {
  printf 'required %s %s, found: %s\n' "$1" "$2" "$3" >&2
  exit 1
}

check_toolchain() {
  local actual

  actual=$(goctl --version 2>&1) || fail_version goctl "$GOCTL_VERSION" "not installed"
  [[ "$actual" == "goctl version $GOCTL_VERSION "* ]] || fail_version goctl "$GOCTL_VERSION" "$actual"

  actual=$(protoc --version 2>&1) || fail_version protoc "$PROTOC_VERSION" "not installed"
  [[ "$actual" == "libprotoc $PROTOC_VERSION" ]] || fail_version protoc "$PROTOC_VERSION" "$actual"

  actual=$(protoc-gen-go --version 2>&1) || fail_version protoc-gen-go "$PROTOC_GEN_GO_VERSION" "not installed"
  [[ "$actual" == "protoc-gen-go v$PROTOC_GEN_GO_VERSION" ]] || fail_version protoc-gen-go "$PROTOC_GEN_GO_VERSION" "$actual"

  actual=$(protoc-gen-go-grpc --version 2>&1) || fail_version protoc-gen-go-grpc "$PROTOC_GEN_GO_GRPC_VERSION" "not installed"
  [[ "$actual" == "protoc-gen-go-grpc $PROTOC_GEN_GO_GRPC_VERSION" ]] || fail_version protoc-gen-go-grpc "$PROTOC_GEN_GO_GRPC_VERSION" "$actual"
}

check_get_form_tags() {
  local api_file=$1
  local types_file=$2
  local request

  while IFS= read -r request; do
    awk -v request="$request" '
      $0 == "type " request " struct {" { in_request = 1; found = 1; next }
      in_request && /^}/ { in_request = 0; exit }
      in_request && /`/ && $0 !~ /`form:"/ { bad = 1 }
      END {
        if (!found) {
          printf "GET request type %s is missing from generated types\n", request > "/dev/stderr"
          exit 1
        }
        if (bad) {
          printf "GET request type %s contains a non-form field tag\n", request > "/dev/stderr"
          exit 1
        }
      }
    ' "$types_file"
  done < <(
    sed -nE 's/^[[:space:]]*get[[:space:]]+[^[:space:]]+[[:space:]]+\(([A-Za-z][A-Za-z0-9_]*)\).*/\1/p' "$api_file"
  )
}

check_toolchain

root=$(git rev-parse --show-toplevel)
temp_root=${TMPDIR:-/tmp}
worktree=$(mktemp -d "${temp_root%/}/hichat-generated-check.XXXXXX")

cleanup() {
  git -C "$root" worktree remove --force "$worktree" >/dev/null 2>&1 || rm -rf "$worktree"
}
trap cleanup EXIT

git -C "$root" worktree add --detach "$worktree" HEAD >/dev/null

if [[ "$committed" == false ]]; then
  for path in "${CONTRACT_FILES[@]}" "${GENERATED_FILES[@]}"; do
    if [[ -f "$root/$path" ]]; then
      mkdir -p "$(dirname "$worktree/$path")"
      cp -p "$root/$path" "$worktree/$path"
    else
      rm -f "$worktree/$path"
    fi
  done
  # The temporary index is the dirty-worktree snapshot baseline. The primary
  # index and worktree are never changed.
  git -C "$worktree" add --all -- "${CONTRACT_FILES[@]}" "${GENERATED_FILES[@]}"
fi

(
  cd "$worktree"
  goctl api go -api apps/social/api/social.api -dir apps/social/api -style gozero >/dev/null
  goctl api go -api apps/im/api/im.api -dir apps/im/api -style gozero >/dev/null
  (cd apps/social/rpc && goctl rpc protoc ./social.proto --go_out=. --go-grpc_out=. --zrpc_out=. >/dev/null)
  (cd apps/im/rpc && goctl rpc protoc ./im.proto --go_out=. --go-grpc_out=. --zrpc_out=. >/dev/null)
)

check_get_form_tags \
  "$worktree/apps/social/api/social.api" \
  "$worktree/apps/social/api/internal/types/types.go"
check_get_form_tags \
  "$worktree/apps/im/api/im.api" \
  "$worktree/apps/im/api/internal/types/types.go"

# Make regenerated files that were deleted from the baseline visible to diff.
git -C "$worktree" add --intent-to-add -- "${GENERATED_FILES[@]}"

if ! git -C "$worktree" diff --quiet -- "${GENERATED_FILES[@]}"; then
  printf '\ngenerated Social/IM API or RPC code is out of date:\n' >&2
  git -C "$worktree" diff -- "${GENERATED_FILES[@]}" >&2
  exit 1
fi

printf 'Social/IM generated code is consistent.\n'
