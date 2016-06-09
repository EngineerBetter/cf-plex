package main

import (
	"bytes"
	"fmt"
	"github.com/EngineerBetter/cf-plex/env"
	"github.com/EngineerBetter/cf-plex/target"
	"github.com/mitchellh/go-homedir"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

var cfUsage = "cf-plex [-g <group>] <cf cli command> [--force]"
var addUsage = "cf-plex add-api [-g <group>] <apiUrl> [<username> <password>]"
var listUsage = "cf-plex list-apis"
var removeUsage = "cf-plex remove-api [-g <group>] <apiUrl>"

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

		if len(args) >= 3 {
			if args[2] == "-g" {
				if len(args) == 5 {
					group := args[3]
					api := args[4]
					fullPath, err := target.AddToGroup(cfPlexHome, group, api)
					bailIfB0rked(err)
					runCf(fullPath, []string{"", "login", "-a", api})
					fmt.Println("Added " + api + " to group '" + group + "'")
					os.Exit(0)
				} else if len(args) == 7 {
					group := args[3]
					api := args[4]
					username := args[5]
					password := args[6]

					fullPath, err := target.AddToGroup(cfPlexHome, group, api)
					bailIfB0rked(err)
					runCf(fullPath, []string{"", "api", api})
					runCf(fullPath, []string{"", "auth", username, password})
					fmt.Println("Added " + api + " to group '" + group + "'")
					os.Exit(0)
				}
			} else if len(args) == 3 {
				api := args[2]
				fullPath, err := target.Add(cfPlexHome, api)
				bailIfB0rked(err)
				runCf(fullPath, []string{"", "login", "-a", api})
				os.Exit(0)
			} else if len(args) == 5 {
				api := args[2]
				username := args[3]
				password := args[4]

				fullPath, err := target.Add(cfPlexHome, api)
				bailIfB0rked(err)
				runCf(fullPath, []string{"", "api", api})
				runCf(fullPath, []string{"", "auth", username, password})
				os.Exit(0)
			}
		}

		fmt.Println("Usage: " + addUsage)
		os.Exit(1)
	case "list-apis":
		bailIfCfEnvs()

		groups, err := target.List(cfPlexHome)
		bailIfB0rked(err)
		for _, group := range groups {
			fmt.Println(group.Name)

			for _, target := range group.Apis {
				fmt.Println("\t" + target.Name)
			}
		}
	case "remove-api":
		bailIfCfEnvs()

		if len(args) < 3 {
			fmt.Println("Usage: " + removeUsage)
			os.Exit(1)
		}

		if args[2] == "-g" {
			if len(args) != 5 {
				fmt.Println("Usage: " + removeUsage)
				os.Exit(1)
			}

			group := args[3]
			api := args[4]
			err := target.RemoveFromGroup(cfPlexHome, group, api)
			bailIfB0rked(err)
			fmt.Println("Removed " + api + " from '" + group + "'")
		} else {
			api := args[2]
			err := target.Remove(cfPlexHome, api)
			bailIfB0rked(err)
			fmt.Println("Removed " + api)
		}
	default:
		var targets []target.Target

		cfEnvs := env.Get("CF_PLEX_APIS", "")
		if cfEnvs != "" {
			targets = getTargetsFromEnv(cfPlexHome, cfEnvs)
		} else {
			if args[1] == "-g" {
				groupName := args[2]
				groups, err := target.List(cfPlexHome)
				bailIfB0rked(err)
				for _, group := range groups {
					if group.Name == groupName {
						targets = group.Apis
					}
				}

				if len(targets) == 0 {
					os.Stderr.WriteString("Group '" + groupName + "' not recognised")
					os.Exit(1)
				}

				args = append(args[0:0], args[2:]...)
			} else {
				if target.GroupsExist(cfPlexHome) {
					os.Stderr.WriteString("-g <group> is mandatory whenever groups have been added. Use '-g default' to target APIs without an explicit group.")
					os.Exit(1)
				}

				var err error
				groups, err := target.List(cfPlexHome)
				bailIfB0rked(err)
				if len(groups[0].Apis) == 0 {
					os.Stderr.WriteString("No APIs have been set")
					os.Exit(1)
				}
				targets = groups[0].Apis
			}
		}

		var force bool
		if args[len(args)-1] == "--force" {
			force = true
			args = args[:len(args)-1]
		}

		fmt.Println()
		for _, aTarget := range targets {
			exitCode, _ := runCf(aTarget.Path, args)
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
		usrHome, err := homedir.Dir()
		bailIfB0rked(err)
		usrHome, err = homedir.Expand(usrHome)
		bailIfB0rked(err)
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

func getTargetsFromEnv(cfPlexHome, cfEnvs string) []target.Target {
	var targets []target.Target
	tripleSeparator := env.Get("CF_PLEX_SEP_TRIPLE", env.PlexTripleSeparator)
	credApiSeparator := env.Get("CF_PLEX_SEP_CREDS_API", env.PlexCredApiSeparator)
	userPassSeparator := env.Get("CF_PLEX_SEP_USER_PASS", env.PlexUserPassSeparator)

	coords, err := env.GetCoordinates(cfEnvs, tripleSeparator, credApiSeparator, userPassSeparator)
	bailIfB0rked(err)

	for _, coord := range coords {
		apiDir, err := target.AddToGroup(cfPlexHome, "batch", coord.Api)
		bailIfB0rked(err)
		targets = append(targets, target.Target{Name: coord.Api, Path: apiDir})

		_, output := runCf(apiDir, []string{"", "api", coord.Api})
		if strings.Contains(output, "Not logged in") {
			runCf(apiDir, []string{"", "auth", coord.Username, coord.Password})
		}
	}

	return targets
}

func runCf(cfHome string, args []string) (int, string) {
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
	bailIfB0rked(err)
	err = cmd.Wait()
	output := buffer.String()
	return determineExitCode(cmd, err), output
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
