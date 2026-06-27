package loftconfig

import (
	"fmt"
	"os/exec"

	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/platform/client"
)

func AuthDevpodCliToPlatform(config *client.Config, logger log.Logger) error {
	cmd := exec.Command("devpod", "pro", "login", "--access-key", config.AccessKey, config.Host)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debugf("Failed executing `devpod pro login`: %v, output: %s", err, out)
		return fmt.Errorf("error executing 'devpod pro login' command: %w, access key: %v, host: %v", err, config.AccessKey, config.Host)
	}

	return nil
}

func AuthVClusterCliToPlatform(config *client.Config, logger log.Logger) error {
	// Check if vcluster is available inside the workspace
	if _, err := exec.LookPath("vcluster"); err != nil {
		logger.Debugf("'vcluster' command is not available")
		return nil
	}

	cmd := exec.Command("vcluster", "login", "--access-key", config.AccessKey, config.Host)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debugf("Failed executing `vcluster login` : %v, output: %s", err, out)
		return fmt.Errorf("error executing 'vcluster login' command: %w", err)
	}

	return nil
}
