package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	devpodhttp "github.com/loft-sh/devpod/pkg/http"
	"github.com/loft-sh/devpod/pkg/inject"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/shell"
	"github.com/loft-sh/devpod/pkg/version"
)

var waitForInstanceConnectionTimeout = time.Minute * 5

func InjectAgent(
	ctx context.Context,
	exec inject.ExecFunc,
	local bool,
	remoteAgentPath,
	downloadURL string,
	preferDownload bool,
	log log.Logger,
	timeout time.Duration,
) error {
	return InjectAgentAndExecute(
		ctx,
		exec,
		local,
		remoteAgentPath,
		downloadURL,
		preferDownload,
		"",
		nil,
		nil,
		nil,
		log,
		timeout,
	)
}

func InjectAgentAndExecute(
	ctx context.Context,
	exec inject.ExecFunc,
	local bool,
	remoteAgentPath,
	downloadURL string,
	preferDownload bool,
	command string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	log log.Logger,
	timeout time.Duration,
) error {
	// should execute locally?
	if local {
		if command == "" {
			return nil
		}

		log.Debugf("Execute command locally")
		return shell.RunEmulatedShell(ctx, command, stdin, stdout, stderr, nil)
	}

	defer log.Debugf("Done InjectAgentAndExecute")
	if remoteAgentPath == "" {
		remoteAgentPath = RemoteDevPodHelperLocation
	}
	if downloadURL == "" {
		downloadURL = DefaultAgentDownloadURL()
	}

	versionCheck := fmt.Sprintf(`[ "$(%s version 2>/dev/null || echo 'false')" != "%s" ]`, remoteAgentPath, version.GetVersion())
	if version.GetVersion() == version.DevVersion {
		preferDownload = false
	}

	// install devpod into the target
	// do a simple hello world to check if we can get something
	now := time.Now()
	lastMessage := time.Now()
	for {
		buf := &bytes.Buffer{}
		if stderr != nil {
			stderr = io.MultiWriter(stderr, buf)
		} else {
			stderr = buf
		}

		scriptParams := &inject.Params{
			Command:             command,
			AgentRemotePath:     remoteAgentPath,
			DownloadURLs:        inject.NewDownloadURLs(downloadURL),
			ExistsCheck:         versionCheck,
			PreferAgentDownload: preferDownload,
			ShouldChmodPath:     true,
		}

		wasExecuted, err := inject.InjectAndExecute(
			ctx,
			exec,
			func(arm bool) (io.ReadCloser, error) {
				return injectBinary(arm, downloadURL, log)
			},
			scriptParams,
			stdin,
			stdout,
			stderr,
			timeout,
			log,
		)
		if err != nil {
			if time.Since(now) > waitForInstanceConnectionTimeout {
				return fmt.Errorf("timeout waiting for instance connection: %w", err)
			} else if wasExecuted {
				return fmt.Errorf("agent error: %s: %w", buf.String(), err)
			}

			if time.Since(lastMessage) > time.Second*5 {
				log.Infof("Waiting for devpod agent to come up...")
				lastMessage = time.Now()
			}

			log.Debugf("Inject Error: %s%v", buf.String(), err)
			time.Sleep(time.Second * 3)
			continue
		}

		break
	}

	return nil
}

func injectBinary(arm bool, tryDownloadURL string, log log.Logger) (io.ReadCloser, error) {
	// this means we need to
	targetArch := "amd64"
	if arm {
		targetArch = "arm64"
	}

	// make sure a linux arm64 binary exists locally
	var err error
	var binaryPath string
	if runtime.GOOS == "linux" && runtime.GOARCH == targetArch {
		binaryPath, err = os.Executable()
		if err != nil {
			return nil, fmt.Errorf("get executable: %w", err)
		}

		// check if we still exist
		_, err = os.Stat(binaryPath)
		if err != nil {
			binaryPath = ""
		}
	}

	// try to look up runner binaries
	if binaryPath == "" {
		binaryPath, err = buildDevAgentBinary(targetArch, log)
		if err != nil {
			return nil, err
		}
	}

	if binaryPath == "" {
		binaryPath = getRunnerBinary(targetArch)
	}

	// download devpod locally
	if binaryPath == "" {
		binaryPath, err = downloadAgentLocally(tryDownloadURL, targetArch, log)
		if err != nil {
			return nil, fmt.Errorf("download agent locally: %w", err)
		}
	}

	// read file
	file, err := os.Open(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("open agent binary: %w", err)
	}

	return file, nil
}

func buildDevAgentBinary(targetArch string, log log.Logger) (string, error) {
	if version.GetVersion() != version.DevVersion {
		return "", nil
	}

	sourceRoot, err := findDevPodSourceRoot()
	if err != nil || sourceRoot == "" {
		return "", err
	}

	agentPath := filepath.Join(os.TempDir(), "devpod-cache", "devpod-linux-"+targetArch)
	if err := os.MkdirAll(filepath.Dir(agentPath), 0755); err != nil {
		return "", fmt.Errorf("create agent path: %w", err)
	}

	tmpPath := agentPath + ".tmp"
	_ = os.Remove(tmpPath)

	var output bytes.Buffer
	cmd := exec.Command("go", "build", "-o", tmpPath, ".")
	cmd.Dir = sourceRoot
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux", "GOARCH="+targetArch, "GOFLAGS=-mod=vendor")
	cmd.Stdout = &output
	cmd.Stderr = &output

	log.Debugf("Building local DevPod agent for linux/%s", targetArch)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build local DevPod agent: %w: %s", err, strings.TrimSpace(output.String()))
	}

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return "", fmt.Errorf("chmod local DevPod agent: %w", err)
	}
	if err := os.Rename(tmpPath, agentPath); err != nil {
		return "", fmt.Errorf("install local DevPod agent: %w", err)
	}

	return agentPath, nil
}

func findDevPodSourceRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return findDevPodSourceRootFrom(wd)
}

func findDevPodSourceRootFrom(dir string) (string, error) {
	for {
		goModPath := filepath.Join(dir, "go.mod")
		rawGoMod, err := os.ReadFile(goModPath)
		if err == nil && strings.Contains(string(rawGoMod), "module github.com/loft-sh/devpod") {
			return dir, nil
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("read %s: %w", goModPath, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func downloadAgentLocally(tryDownloadURL, targetArch string, log log.Logger) (string, error) {
	agentPath := filepath.Join(os.TempDir(), "devpod-cache", "devpod-linux-"+targetArch)
	err := os.MkdirAll(filepath.Dir(agentPath), 0755)
	if err != nil {
		return "", fmt.Errorf("create agent path: %w", err)
	}

	stat, statErr := os.Stat(agentPath)
	if version.GetVersion() == version.DevVersion && statErr == nil {
		return agentPath, nil
	}

	fullDownloadURL := tryDownloadURL + "/devpod-linux-" + targetArch
	log.Debugf("Attempting to download DevPod agent from: %s", fullDownloadURL)

	resp, err := devpodhttp.GetHTTPClient().Get(fullDownloadURL)
	if err != nil {
		return "", fmt.Errorf("download devpod: %w", err)
	}
	defer resp.Body.Close()

	if statErr == nil && stat.Size() == resp.ContentLength {
		return agentPath, nil
	}

	log.Infof("Download DevPod Agent...")
	file, err := os.Create(agentPath)
	if err != nil {
		return "", fmt.Errorf("create agent binary: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		_ = os.Remove(agentPath)
		return "", fmt.Errorf("failed to download devpod from URL %s: %w", fullDownloadURL, err)
	}

	return agentPath, nil
}

func getRunnerBinary(targetArch string) string {
	binaryPath := filepath.Join(os.TempDir(), "devpod-cache", "devpod-linux-"+targetArch)
	_, err := os.Stat(binaryPath)
	if err != nil {
		return ""
	}
	return binaryPath
}
