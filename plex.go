package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	args := os.Args

	cfPlexHome := getConfigDir()
	runCf(cfPlexHome, args)
}

func SetEnv(key, value string, env []string) []string {
	var indexOfKey int
	var found bool

	for index, pair := range env {
		if strings.HasPrefix(pair, key+"=") {
			found = true
			indexOfKey = index
		}
	}

	env = append(env[:])
	if found {
		env = append(env[:indexOfKey], env[indexOfKey+1:]...)
	}

	return append(env, key+"="+value)
}

func CommandWithEnv(env []string, args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	return cmd
}

func Output(cmd *exec.Cmd) string {
	bytes, err := cmd.Output()
	if err != nil {
		os.Exit(1)
	}
	return string(bytes)
}

func getConfigDir() (configDir string) {
	configDir = os.Getenv("CF_PLEX_HOME")
	if configDir == "" {
		usr, err := user.Current()
		bailIfB0rked(err)
		usrHome := usr.HomeDir
		configDir = filepath.Join(usrHome, ".cfplex")
	}
	return
}

func runCf(cfPlexHome string, args []string) {
	args[0] = "cf"
	env := SetEnv("CF_HOME", cfPlexHome, os.Environ())
	cmd := CommandWithEnv(env, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	err := cmd.Wait()
	os.Exit(determineExitCode(cmd, err))
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

func bailIfB0rked(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
