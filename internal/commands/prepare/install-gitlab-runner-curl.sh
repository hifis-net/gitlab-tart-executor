#!/bin/bash

# Set shell options to enable fail-fast behavior
#
# * -e: fail the script when an error occurs or command fails
# * -u: fail the script when attempting to reference unset parameters
# * -o pipefail: by default an exit status of a pipeline is that of its
#                last command, this fails the pipe early if an error in
#                any of its commands occurs
#
set -euo pipefail

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  aarch64)
      ARCH="arm64"
      ;;
  x86_64)
      ARCH="amd64"
      ;;
esac

GITLAB_RUNNER_VERSION="latest"
GITLAB_RUNNER_URL="https://gitlab-runner-downloads.s3.amazonaws.com/${GITLAB_RUNNER_VERSION}/binaries/gitlab-runner-${OS}-${ARCH}"

# Replace the GitLab Runner that is currently resolved via PATH (for example the
# one that comes preinstalled with the image) instead of adding a second binary
# that might not be the first one in PATH. Whether an installation is needed at
# all is decided by the caller.
GITLAB_RUNNER_PATH="$(command -v gitlab-runner || echo "/usr/local/bin/gitlab-runner")"

echo "Installing GitLab Runner using cURL..."

if ! sudo -n true &>/dev/null
then
  echo "Failed to install GitLab Runner using cURL: passwordless sudo is required, but is not configured"

  exit 1
fi

TEMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TEMP_DIR}"' EXIT

echo "Downloading GitLab Runner from $GITLAB_RUNNER_URL..."
curl --fail --silent --show-error --location --output "${TEMP_DIR}/gitlab-runner" "$GITLAB_RUNNER_URL"

echo "Making the downloaded GitLab Runner executable..."
chmod +x "${TEMP_DIR}/gitlab-runner"

# Install atomically, so that a failed download never leaves a truncated
# binary behind. Note that replacing a symlink that points to a Homebrew
# installation only replaces the symlink and leaves the Homebrew keg intact.
echo "Installing GitLab Runner to $GITLAB_RUNNER_PATH..."
sudo mkdir -p "$(dirname "$GITLAB_RUNNER_PATH")"
sudo mv -f "${TEMP_DIR}/gitlab-runner" "$GITLAB_RUNNER_PATH"

echo "GitLab Runner was successfully installed!"
