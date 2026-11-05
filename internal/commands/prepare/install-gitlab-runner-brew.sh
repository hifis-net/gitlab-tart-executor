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

# Whether GitLab Runner needs to be installed is decided by the caller,
# so always install or upgrade it here.
echo "Updating Homebrew..."

brew update

echo "Installing GitLab Runner via Homebrew..."

if brew list --versions gitlab-runner &> /dev/null
then
  brew upgrade gitlab-runner
else
  brew install gitlab-runner
fi

echo "GitLab Runner was successfully installed!"
