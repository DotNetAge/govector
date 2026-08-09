#!/bin/bash
# =============================================================================
# GoVector Homebrew Formula Generator
# =============================================================================
# Downloads darwin binaries from GitHub Release, computes SHA256,
# generates Formula/govector.rb, and pushes to homebrew-govector tap.
#
# Usage:
#   ./scripts/homebrew-release.sh [VERSION]
#   VERSION defaults to latest git tag if not provided
# =============================================================================

set -euo pipefail

# ── Configuration ──
GITHUB_REPO="DotNetAge/govector"
TAP_REPO="DotNetAge/homebrew-govector"
FORMULA_NAME="govector"
TEMPLATE="scripts/release/govector.rb"
OUTPUT_DIR="dist/homebrew"

# ── Version ──
VERSION="${1:-$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')}"
if [ -z "${VERSION}" ] || [ "${VERSION}" = "dev" ]; then
  echo "ERROR: Cannot determine version. Pass VERSION as argument or run from a tagged commit."
  exit 1
fi

TAG="v${VERSION}"
echo "=== GoVector Homebrew Formula Generator ==="
echo "Version: ${VERSION} (tag: ${TAG})"

# ── Prepare output ──
mkdir -p "${OUTPUT_DIR}"
FORMULA="${OUTPUT_DIR}/${FORMULA_NAME}.rb"

# ── Download & compute SHA256 for each platform ──
download_sha256() {
  local url="$1"
  local filename="$2"
  echo "  Downloading: ${url} → ${filename}"
  curl -sL "${url}" -o "${filename}"
  shasum -a 256 "${filename}" | awk '{print $1}'
}

ARM64_URL="https://github.com/${GITHUB_REPO}/releases/download/${TAG}/govector_v${VERSION}_darwin_arm64.tar.gz"
AMD64_URL="https://github.com/${GITHUB_REPO}/releases/download/${TAG}/govector_v${VERSION}_darwin_amd64.tar.gz"
LINUX_AMD64_URL="https://github.com/${GITHUB_REPO}/releases/download/${TAG}/govector_v${VERSION}_linux_amd64.tar.gz"

ARM64_SHA=$(download_sha256 "${ARM64_URL}" "/tmp/govector_darwin_arm64.tar.gz")
AMD64_SHA=$(download_sha256 "${AMD64_URL}" "/tmp/govector_darwin_amd64.tar.gz")
LINUX_AMD64_SHA=$(download_sha256 "${LINUX_AMD64_URL}" "/tmp/govector_linux_amd64.tar.gz")

echo "  SHA256 (darwin/arm64): ${ARM64_SHA}"
echo "  SHA256 (darwin/amd64): ${AMD64_SHA}"
echo "  SHA256 (linux/amd64):  ${LINUX_AMD64_SHA}"

# ── Generate formula from template ──
sed \
  -e "s|__VERSION__|${VERSION}|g" \
  -e "s|__SHA256_ARM64__|${ARM64_SHA}|g" \
  -e "s|__SHA256_AMD64__|${AMD64_SHA}|g" \
  -e "s|__SHA256_LINUX__|${LINUX_AMD64_SHA}|g" \
  -e "s|__GITHUB_REPO__|${GITHUB_REPO}|g" \
  "${TEMPLATE}" > "${FORMULA}"

echo ""
echo "=== Generated: ${FORMULA} ==="
cat "${FORMULA}"

# ── Push to tap repo (only in CI) ──
if [ -n "${GITHUB_TOKEN:-}" ] && [ -n "${HOMEBREW_TAP_TOKEN:-}" ]; then
  TOKEN="${HOMEBREW_TAP_TOKEN:-${GITHUB_TOKEN}}"
  TAP_DIR="/tmp/homebrew-${FORMULA_NAME}-tap"

  echo ""
  echo "=== Pushing to ${TAP_REPO} ==="

  # Clone or use existing clone
  if [ -d "${TAP_DIR}" ]; then
    cd "${TAP_DIR}" && git pull origin main 2>/dev/null || true
  else
    git clone "https://x-access-token:${TOKEN}@github.com/${TAP_REPO}.git" "${TAP_DIR}"
  fi

  mkdir -p "${TAP_DIR}/Formula"
  cp "${FORMULA}" "${TAP_DIR}/Formula/${FORMULA_NAME}.rb"

  cd "${TAP_DIR}"
  git config user.name "github-actions[bot]"
  git config user.email "github-actions[bot]@users.noreply.github.com"

  if git diff --quiet; then
    echo "No changes to push (formula unchanged)"
  else
    git add "Formula/${FORMULA_NAME}.rb"
    git commit -m "${FORMULA_NAME} ${VERSION}"
    git push origin HEAD:main 2>&1 || {
      echo "WARNING: Could not push to main, trying master..."
      git push origin HEAD:master 2>&1 || echo "WARNING: Could not push to tap repo."
    }
    echo "✅ Pushed ${FORMULA_NAME} ${VERSION} to ${TAP_REPO}"
  fi
else
  echo ""
  echo "NOTE: GITHUB_TOKEN / HOMEBREW_TAP_TOKEN not set. Skipping push."
  echo "      Formula saved at: ${FORMULA}"
  echo "      To publish manually:"
  echo "        cp ${FORMULA} <tap-repo>/Formula/${FORMULA_NAME}.rb"
fi
