#!/usr/bin/env bash
# deploy/scripts/lib/build.sh — compile and verify binary

run_build() {
  phase_start "BUILD"

  local arch="${BUILD_ARCH:-arm64}"
  info "Building for linux/$arch..."

  GOOS=linux GOARCH="$arch" go build -o dist/mathsvr ./cmd/mathsvr \
    || fatal "Build failed"

  # Verify binary exists and is non-empty
  [[ -s dist/mathsvr ]] || fatal "Binary dist/mathsvr is empty or missing"

  local size
  size=$(wc -c < dist/mathsvr | tr -d ' ')
  info "Binary size: ${size} bytes"

  # Verify it's the right architecture
  local file_type
  file_type=$(file dist/mathsvr)
  info "Binary type: $file_type"

  if [[ "$arch" == "arm64" ]] && ! echo "$file_type" | grep -qi "aarch64\|arm64"; then
    fatal "Binary architecture mismatch — expected arm64"
  fi

  if [[ "$arch" == "amd64" ]] && ! echo "$file_type" | grep -qi "x86-64\|amd64"; then
    fatal "Binary architecture mismatch — expected amd64"
  fi

  phase_end "BUILD"
}
