//go:build !windows

package config

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func PrepareCmdUser(cmd *exec.Cmd, userName string) error {
	// execute as user
	u, err := user.Lookup(userName)
	if err != nil {
		return fmt.Errorf("lookup user %s: %w", userName, err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("parse uid %s: %w", u.Uid, err)
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("parse gid %s: %w", u.Gid, err)
	}
	groupIDs, err := getUserGroupIDs(u)
	if err != nil {
		return err
	}
	cmd.Env = patchEnvVars(cmd.Environ(), map[string]string{
		"HOME":    u.HomeDir,
		"USER":    u.Username,
		"LOGNAME": u.Username,
	})

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:    uint32(uid),
			Gid:    uint32(gid),
			Groups: groupIDs,
		},
	}

	return nil
}

func getUserGroupIDs(u *user.User) ([]uint32, error) {
	groups, err := u.GroupIds()
	if err != nil {
		return nil, fmt.Errorf("lookup groups for user %s: %w", u.Username, err)
	}

	groupIDs := make([]uint32, 0, len(groups))
	for _, group := range groups {
		groupID, err := strconv.Atoi(group)
		if err != nil {
			return nil, fmt.Errorf("parse group id %s: %w", group, err)
		}
		groupIDs = append(groupIDs, uint32(groupID))
	}

	return groupIDs, nil
}

func patchEnvVars(env []string, patches map[string]string) []string {
	newEnv := map[string]string{}
	for _, v := range env {
		key, value, ok := strings.Cut(v, "=")
		if !ok {
			continue
		}
		newEnv[key] = value
	}

	// apply patches
	for k, v := range patches {
		newEnv[k] = v
	}

	retEnv := []string{}
	for k, v := range newEnv {
		retEnv = append(retEnv, fmt.Sprintf("%s=%s", k, v))
	}

	return retEnv
}
