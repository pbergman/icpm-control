package main

import (
	"context"
	"fmt"
	"github.com/pbergman/logger"
	"io"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func getShellFromScript(script *Script) (string, []string) {
	var shell = "bash"
	var args []string
	if "" != script.Shell {
		var parts = strings.Split(script.Shell, " ")
		shell = parts[0]
		if len(parts) > 1 {
			args = parts[1:]
		}

	}
	return shell, args
}

func StartShell(ctx context.Context, script *Script, logger *logger.Logger) (*exec.Cmd, io.WriteCloser) {

	shell, args := getShellFromScript(script)

	logger.Debug(fmt.Sprintf("exec: shell: %s args: %q", shell, args))

	var cmd = exec.CommandContext(ctx, shell, args...)
	cmd.Stdout = &LogWriter{Format: "exec.stdout> %s", Logger: logger.Debug}
	cmd.Stderr = &LogWriter{Format: "exec.stderr> %s", Logger: logger.Error}

	if script.User != "" || script.Group != "" {
		var attr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{},
		}

		if script.Group != "" {
			if val, err := strconv.Atoi(script.Group); err == nil {
				attr.Credential.Gid = uint32(val)
			} else {
				grp, err := user.LookupGroup(script.Group)
				if err != nil {
					logger.Error(fmt.Sprintf("could not find group %s", script.Group))
				} else {
					uid, _ := strconv.Atoi(grp.Gid)
					attr.Credential.Gid = uint32(uid)
				}
			}
		}

		if script.User != "" {
			if val, err := strconv.Atoi(script.User); err == nil {
				attr.Credential.Uid = uint32(val)
			} else {
				usr, err := user.Lookup(script.User)

				if err != nil {
					logger.Error(fmt.Sprintf("could not find user %s", script.User))
				} else {
					uid, _ := strconv.Atoi(usr.Uid)
					attr.Credential.Uid = uint32(uid)
				}
			}
		}

		cmd.SysProcAttr = attr
		logger.Debug(fmt.Sprintf("exec: uuid: %d guid: %d", cmd.SysProcAttr.Credential.Uid, cmd.SysProcAttr.Credential.Gid))
	}

	in, err := cmd.StdinPipe()

	if err != nil {
		logger.Error(err)
		return nil, nil
	}

	if err := cmd.Start(); err != nil {
		logger.Error(err)
		return nil, nil
	}

	return cmd, in
}
