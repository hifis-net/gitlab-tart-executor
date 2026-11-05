package prepare

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"golang.org/x/crypto/ssh"
)

// gitlabRunnerVersionVariable is the version of the GitLab Runner that runs
// this job and thus generates the job script that is executed inside the VM,
// for example "gitlab-runner artifacts-uploader --timeout".
//
// The GitLab Runner exposes this variable as "CUSTOM_ENV_CI_RUNNER_VERSION"
// to the custom executor's stages.
const gitlabRunnerVersionVariable = "CUSTOM_ENV_CI_RUNNER_VERSION"

// gitlabRunnerVersion returns the version of the GitLab Runner that is
// available in the VM or an empty string when there is none.
func gitlabRunnerVersion(client *ssh.Client) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout bytes.Buffer

	session.Stdout = &stdout
	session.Stderr = io.Discard

	if err := session.Run("gitlab-runner --version"); err != nil {
		var exitError *ssh.ExitError
		if errors.As(err, &exitError) {
			// The "gitlab-runner" command is either missing or not functional,
			// in both cases it makes sense to (re-)install it
			return "", nil
		}

		return "", err
	}

	for line := range strings.SplitSeq(stdout.String(), "\n") {
		version, found := strings.CutPrefix(strings.TrimSpace(line), "Version:")
		if found {
			return strings.TrimSpace(version), nil
		}
	}

	// The output doesn't look like the one produced by a proper GitLab Runner
	return "", nil
}

// gitlabRunnerInstallNeeded tells whether the GitLab Runner found in the VM
// satisfies the configured installation mode and, if it doesn't, why.
func gitlabRunnerInstallNeeded(mode string, installedVersion string) (bool, string) {
	if installedVersion == "" {
		return true, "there's no GitLab Runner installed in the VM"
	}

	installed, err := semver.NewVersion(installedVersion)
	if err != nil {
		return true, fmt.Sprintf("the installed version %q could not be parsed", installedVersion)
	}

	switch mode {
	case "brew", "curl", "true", "yes", "on":
		// These modes install the latest version available to Homebrew or via
		// cURL, so the best we can do is to make sure that the installed
		// version is not older than the one that generated the job script.
		//
		// See https://github.com/cirruslabs/gitlab-tart-executor/issues/143
		hostVersion := os.Getenv(gitlabRunnerVersionVariable)

		host, err := semver.NewVersion(hostVersion)
		if hostVersion == "" || err != nil {
			// We can't tell whether the installed version is adequate,
			// so don't touch it and preserve the old behavior
			return false, ""
		}

		if installed.LessThan(host) {
			return true, fmt.Sprintf("the installed version %s is older than the "+
				"GitLab Runner %s that generated the job script", installedVersion, hostVersion)
		}

		return false, ""
	default:
		// An explicit "major.minor.patch" version was requested
		requested, err := semver.NewVersion(mode)
		if err != nil {
			return false, ""
		}

		if !installed.Equal(requested) {
			return true, fmt.Sprintf("the installed version %s does not match "+
				"the requested version %s", installedVersion, requested.String())
		}

		return false, ""
	}
}

// installGitlabRunner transfers the install script to the VM and executes it.
func installGitlabRunner(client *ssh.Client, script string) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdinBuf, err := session.StdinPipe()
	if err != nil {
		return err
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		return err
	}

	// Expose GitLab job environment variables to GitLab Runner installation script
	for _, keyAndValue := range os.Environ() {
		keyAndValue, found := strings.CutPrefix(keyAndValue, "CUSTOM_ENV_")
		if !found {
			// Doesn't look like a GitLab job environment variable
			continue
		}

		key, value, _ := strings.Cut(keyAndValue, "=")

		if _, err := fmt.Fprintf(stdinBuf, "export %s=%s\n", key, quote(value)); err != nil {
			return err
		}
	}

	if _, err := stdinBuf.Write([]byte(script)); err != nil {
		return err
	}

	if err := stdinBuf.Close(); err != nil {
		return err
	}

	return session.Wait()
}

func logGitlabRunnerVersionMismatch(mode string, installedVersion string) {
	if needed, reason := gitlabRunnerInstallNeeded(mode, installedVersion); needed {
		log.Printf("WARNING: the GitLab Runner in the VM still doesn't satisfy the requested "+
			"installation mode: %s\n", reason)
	}
}
