#!/usr/bin/env bash
# Syncs this minimal fork from upstream opentelemetry-collector-contrib.
# Only env + lambda detectors are kept; all other detector code is not fetched.
#
# Usage:
#   ./sync-from-upstream.sh v0.146.0
#
# Prerequisites: gh CLI authenticated, patch, go

set -euo pipefail

VERSION="${1:?Usage: $0 <version>  e.g. $0 v0.146.0}"
UPSTREAM="open-telemetry/opentelemetry-collector-contrib"
REF="$VERSION"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

fetch() {
  local path="$1" dest="$2"
  mkdir -p "$(dirname "$dest")"
  echo "  fetching $path"
  gh api "repos/${UPSTREAM}/contents/${path}?ref=${REF}" --jq '.content' | base64 -d > "$dest"
}

echo "==> Syncing resourcedetectionprocessor from upstream $VERSION"

# ── verbatim files (no local changes) ────────────────────────────────────────
fetch processor/resourcedetectionprocessor/resourcedetection_processor.go \
      "$SCRIPT_DIR/resourcedetection_processor.go"

fetch processor/resourcedetectionprocessor/internal/resourcedetection.go \
      "$SCRIPT_DIR/internal/resourcedetection.go"

fetch processor/resourcedetectionprocessor/internal/context.go \
      "$SCRIPT_DIR/internal/context.go"

fetch processor/resourcedetectionprocessor/internal/metadata/generated_status.go \
      "$SCRIPT_DIR/internal/metadata/generated_status.go"

fetch processor/resourcedetectionprocessor/internal/metadata/generated_feature_gates.go \
      "$SCRIPT_DIR/internal/metadata/generated_feature_gates.go"

fetch processor/resourcedetectionprocessor/internal/env/env.go \
      "$SCRIPT_DIR/internal/env/env.go"

fetch processor/resourcedetectionprocessor/internal/aws/lambda/lambda.go \
      "$SCRIPT_DIR/internal/aws/lambda/lambda.go"

fetch processor/resourcedetectionprocessor/internal/aws/lambda/config.go \
      "$SCRIPT_DIR/internal/aws/lambda/config.go"

fetch processor/resourcedetectionprocessor/internal/aws/lambda/internal/metadata/generated_config.go \
      "$SCRIPT_DIR/internal/aws/lambda/internal/metadata/generated_config.go"

fetch processor/resourcedetectionprocessor/internal/aws/lambda/internal/metadata/generated_resource.go \
      "$SCRIPT_DIR/internal/aws/lambda/internal/metadata/generated_resource.go"

# ── patched files (fetch upstream, apply local patch) ────────────────────────
echo "  patching factory.go"
fetch processor/resourcedetectionprocessor/factory.go /tmp/rdp_factory_upstream.go
patch -o "$SCRIPT_DIR/factory.go" /tmp/rdp_factory_upstream.go "$SCRIPT_DIR/factory.patch" \
  || { echo "ERROR: factory.patch failed to apply — see above. Update factory.patch for $VERSION."; exit 1; }

echo "  patching config.go"
fetch processor/resourcedetectionprocessor/config.go /tmp/rdp_config_upstream.go
patch -o "$SCRIPT_DIR/config.go" /tmp/rdp_config_upstream.go "$SCRIPT_DIR/config.patch" \
  || { echo "ERROR: config.patch failed to apply — see above. Update config.patch for $VERSION."; exit 1; }

# ── update go.mod version annotation ─────────────────────────────────────────
echo "  updating go.mod go version annotation"
UPSTREAM_GO_VER=$(gh api "repos/${UPSTREAM}/contents/processor/resourcedetectionprocessor/go.mod?ref=${REF}" \
  --jq '.content' | base64 -d | awk '/^go /{print $2}')
sed -i.bak "s/^go .*/go ${UPSTREAM_GO_VER}/" "$SCRIPT_DIR/go.mod" && rm "$SCRIPT_DIR/go.mod.bak"

# ── tidy ─────────────────────────────────────────────────────────────────────
echo "  go mod tidy (fork)"
(cd "$SCRIPT_DIR" && go mod tidy)

echo "  go mod tidy (lambdacomponents)"
(cd "$SCRIPT_DIR/../.." && go mod tidy)

echo ""
echo "Done. Review changes with: git diff"
echo "If the patches failed to apply cleanly, regenerate them:"
echo "  diff -u /tmp/rdp_factory_upstream.go factory.go > factory.patch"
echo "  diff -u /tmp/rdp_config_upstream.go  config.go  > config.patch"
