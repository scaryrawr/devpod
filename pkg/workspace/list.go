package workspace

import (
	"context"
	"os"
	"strings"

	"github.com/loft-sh/devpod/pkg/config"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/platform"
	providerpkg "github.com/loft-sh/devpod/pkg/provider"
)

func List(ctx context.Context, devPodConfig *config.Config, skipPro bool, owner platform.OwnerFilter, log log.Logger) ([]*providerpkg.Workspace, error) {
	localWorkspaces, err := ListLocalWorkspaces(devPodConfig.DefaultContext, log)
	if err != nil {
		return nil, err
	}

	workspaces := map[string]*providerpkg.Workspace{}
	for _, workspace := range localWorkspaces {
		workspaces[workspace.UID] = workspace
	}

	retWorkspaces := []*providerpkg.Workspace{}
	for _, v := range workspaces {
		retWorkspaces = append(retWorkspaces, v)
	}

	return retWorkspaces, nil
}

func ListLocalWorkspaces(contextName string, log log.Logger) ([]*providerpkg.Workspace, error) {
	workspaceDir, err := providerpkg.GetWorkspacesDir(contextName)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(workspaceDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	retWorkspaces := []*providerpkg.Workspace{}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		workspaceConfig, err := providerpkg.LoadWorkspaceConfig(contextName, entry.Name())
		if err != nil {
			log.ErrorStreamOnly().Warnf("Couldn't load workspace %s: %v", entry.Name(), err)
			continue
		}

		if workspaceConfig.IsPro() {
			continue
		}

		retWorkspaces = append(retWorkspaces, workspaceConfig)
	}

	return retWorkspaces, nil
}
