package main

import (
	"bytes"
	"fmt"
	"github.com/EngineerBetter/cf-plex/env"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
)

var cfUsage = "cf-plex <cf cli command> [--force]"
var addUsage = "cf-plex add-api <apiUrl> [<username> <password>]"
var listUsage = "cf-plex list-apis"
var removeUsage = "cf-plex remove-api <apiUrl>"

func main() {
	args := os.Args
	cfPlexHome := getConfigDir()

	if len(args) == 1 {
		printUsageAndBail()
	}

	switch args[1] {
	case "help":
	case "--help":
		printUsageAndBail()
	case "add-api":
		bailIfCfEnvs()

		switch len(args) {
		case 3:
			api := args[2]
			apiDir := sanitiseApi(api)
			fullPath := filepath.Join(cfPlexHome, apiDir)
			err := os.MkdirAll(fullPath, 0700)
			bailIfB0rked(err)
			runCf(fullPath, []string{"", "login", "-a", api})
		case 5:
			api := args[2]
			username := args[3]
			password := args[4]

			apiDir := sanitiseApi(api)
			fullPath := filepath.Join(cfPlexHome, apiDir)
			err := os.MkdirAll(fullPath, 0700)
			bailIfB0rked(err)
			runCf(fullPath, []string{"", "api", api})
			runCf(fullPath, []string{"", "auth", username, password})
		default:
			fmt.Println("Usage: " + addUsage)
			os.Exit(1)
		}
	case "list-apis":
		bailIfCfEnvs()

		apiDirs, err := getApiDirs(cfPlexHome)
		bailIfB0rked(err)
		for _, apiDir := range apiDirs {
			fmt.Println(path.Base(apiDir))
		}
	case "remove-api":
		bailIfCfEnvs()

		if len(args) < 3 {
			fmt.Println("Usage: " + removeUsage)
			os.Exit(1)
		}

		api := args[2]
		apiDir := sanitiseApi(api)
		fullPath := filepath.Join(cfPlexHome, apiDir)
		err := os.RemoveAll(fullPath)
		bailIfB0rked(err)
		fmt.Println("Removed " + api)
	default:
		var apiDirs []string

		cfEnvs := env.Get("CF_PLEX_APIS", "")
		if cfEnvs != "" {
			tripleSeparator := env.Get("CF_PLEX_SEP_TRIPLE", env.PlexTripleSeparator)
			credApiSeparator := env.Get("CF_PLEX_SEP_CREDS_API", env.PlexCredApiSeparator)
			userPassSeparator := env.Get("CF_PLEX_SEP_USER_PASS", env.PlexUserPassSeparator)

			coords, err := env.GetCoordinates(cfEnvs, tripleSeparator, credApiSeparator, userPassSeparator)
			bailIfB0rked(err)

			for _, coord := range coords {
				apiSanitised := sanitiseApi(coord.Api)
				apiDir := filepath.Join(cfPlexHome, "batch", apiSanitised)
				os.MkdirAll(apiDir, 0700)
				apiDirs = append(apiDirs, apiDir)

				_, output := runCf(apiDir, []string{"", "api", coord.Api})
				if strings.Contains(output, "Not logged in") {
					runCf(apiDir, []string{"", "auth", coord.Username, coord.Password})
				}
			}
		} else {
			var err error
			apiDirs, err = getApiDirs(cfPlexHome)
			bailIfB0rked(err)
			if len(apiDirs) == 0 {
				os.Stderr.WriteString("No APIs have been set")
				os.Exit(1)
			}
		}

		var force bool
		if args[len(args)-1] == "--force" {
			force = true
			args = args[:len(args)-1]
		}

		for _, apiDir := range apiDirs {
			exitCode, _ := runCf(apiDir, args)
			if exitCode != 0 && !force {
				os.Exit(exitCode)
			}
		}
	}
}

func CommandWithEnv(env []string, args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	return cmd
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

func bailIfCfEnvs() {
	if env.Get("CF_PLEX_APIS", "") != "" {
		fmt.Println("Managing APIs is not allowed when CF_PLEX_APIS is set")
		os.Exit(1)
	}
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

func runCf(cfHome string, args []string) (int, string) {
	args[0] = "cf"
	env := env.Set("CF_HOME", cfHome, os.Environ())
	cmd := CommandWithEnv(env, args...)

	var buffer bytes.Buffer
	multiWriter := io.MultiWriter(os.Stdout, &buffer)

	cmd.Stdin = os.Stdin
	cmd.Stdout = multiWriter
	cmd.Stderr = os.Stderr

	fmt.Printf("Running '%s' on %s\n", strings.Join(args, " "), path.Base(cfHome))
	cmd.Start()
	err := cmd.Wait()
	return determineExitCode(cmd, err), buffer.String()
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

func printUsageAndBail() {
	fmt.Println("Usage:")
	fmt.Println(cfUsage)
	fmt.Println(addUsage)
	fmt.Println(listUsage)
	fmt.Println(removeUsage)
	os.Exit(1)
}

func bailIfB0rked(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
