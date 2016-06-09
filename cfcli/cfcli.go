package cfcli

import (
	"github.com/EngineerBetter/cf-plex/env"

	"bytes"

	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

func CommandWithEnv(env []string, args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	return cmd
}

func Run(cfHome string, args []string) (error, int, string) {
	args[0] = "cf"
	env := env.Set("CF_HOME", cfHome, os.Environ())
	cmd := CommandWithEnv(env, args...)

	buffer := bytes.NewBufferString("")
	multiWriter := io.MultiWriter(os.Stdout, buffer)

	cmd.Stdin = os.Stdin
	cmd.Stdout = multiWriter
	cmd.Stderr = os.Stderr

	status := fmt.Sprintf("\nRunning '%s' on %s\n", strings.Join(args, " "), path.Base(cfHome))

	if args[1] == "auth" {
		status = strings.Replace(status, args[3], "[expunged]", -1)
	}

	fmt.Printf(status)
	err := cmd.Start()

	if err != nil {
		return err, -1, ""
	}

	err = cmd.Wait()
	output := buffer.String()
	return nil, determineExitCode(cmd, err), output
}

func determineExitCode(cmd *exec.Cmd, err error) (exitCode int) {
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if status.Signaled() {
		exitCode = 128 + int(status.Signal())
	} else {
		exitStatus := status.ExitStatus()
		if exitStatus == -1 && err != nil {
			exitCode = 254
		}
		exitCode = exitStatus
	}

	return
}
