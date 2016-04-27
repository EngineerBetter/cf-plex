package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

func main() {
	args := os.Args
	cfPlexHome := getConfigDir()

	switch args[1] {
	case "add-api":
		api := args[2]
		username := args[3]
		password := args[4]

		apiDir := sanitiseApi(api)
		fullPath := filepath.Join(cfPlexHome, apiDir)
		err := os.MkdirAll(fullPath, 0700)
		bailIfB0rked(err)
		runCf(fullPath, []string{"", "api", api})
		runCf(fullPath, []string{"", "auth", username, password})
	case "list-apis":
		apiDirs, err := getApiDirs(cfPlexHome)
		bailIfB0rked(err)
		for _, apiDir := range apiDirs {
			fmt.Println(apiDir)
		}
	default:
		apiDirs, err := getApiDirs(cfPlexHome)
		bailIfB0rked(err)
		if len(apiDirs) == 0 {
			os.Stderr.WriteString("No APIs have been set")
			os.Exit(1)
		}
		for _, apiDir := range apiDirs {
			runCf(apiDir, args)
		}
	}
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

func sanitiseApi(api string) string {
	api = strings.Replace(api, ":", "_", -1)
	api = strings.Replace(api, "/", "_", -1)
	return api
}

func getApiDirs(configDir string) ([]string, error) {
	f, err := os.Open(configDir)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	for index, apiDir := range names {
		names[index] = filepath.Join(configDir, apiDir)
	}

	return names, nil
}

func runCf(cfHome string, args []string) {
	args[0] = "cf"
	env := SetEnv("CF_HOME", cfHome, os.Environ())
	cmd := CommandWithEnv(env, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	err := cmd.Wait()
	exitCode := determineExitCode(cmd, err)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
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
