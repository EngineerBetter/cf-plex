package main

import (
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := os.Args
	args[0] = "cf"
	cmd := CommandWithEnv(os.Environ(), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
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
