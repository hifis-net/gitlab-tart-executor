package prepare

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGitlabRunnerInstallNeeded(t *testing.T) {
	testCases := []struct {
		name             string
		mode             string
		hostVersion      string
		installedVersion string
		expectedNeeded   bool
	}{
		{
			name:             "not installed",
			mode:             "brew",
			hostVersion:      "19.5.2",
			installedVersion: "",
			expectedNeeded:   true,
		},
		{
			name:             "brew: older than the host runner",
			mode:             "brew",
			hostVersion:      "19.5.2",
			installedVersion: "18.4.0",
			expectedNeeded:   true,
		},
		{
			name:             "brew: same as the host runner",
			mode:             "brew",
			hostVersion:      "19.5.2",
			installedVersion: "19.5.2",
			expectedNeeded:   false,
		},
		{
			name:             "brew: newer than the host runner",
			mode:             "brew",
			hostVersion:      "19.5.2",
			installedVersion: "19.6.0",
			expectedNeeded:   false,
		},
		{
			name:             "brew: unknown host runner version",
			mode:             "brew",
			hostVersion:      "",
			installedVersion: "18.4.0",
			expectedNeeded:   false,
		},
		{
			name:             "curl: older than the host runner",
			mode:             "curl",
			hostVersion:      "19.5.2",
			installedVersion: "18.4.0",
			expectedNeeded:   true,
		},
		{
			name:             "legacy auto mode",
			mode:             "true",
			hostVersion:      "19.5.2",
			installedVersion: "18.4.0",
			expectedNeeded:   true,
		},
		{
			name:             "explicit version: mismatch",
			mode:             "18.9.0",
			installedVersion: "18.4.0",
			expectedNeeded:   true,
		},
		{
			name:             "explicit version: match",
			mode:             "18.9.0",
			installedVersion: "18.9.0",
			expectedNeeded:   false,
		},
		{
			name:             "explicit version: newer installed version",
			mode:             "18.9.0",
			installedVersion: "19.3.3",
			expectedNeeded:   true,
		},
		{
			name:             "unparseable installed version",
			mode:             "brew",
			hostVersion:      "19.5.2",
			installedVersion: "development version",
			expectedNeeded:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(gitlabRunnerVersionVariable, testCase.hostVersion)

			needed, reason := gitlabRunnerInstallNeeded(testCase.mode, testCase.installedVersion)

			require.Equal(t, testCase.expectedNeeded, needed)
			if testCase.expectedNeeded {
				require.NotEmpty(t, reason)
			}
		})
	}
}
