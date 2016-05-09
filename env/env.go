package env

import (
	"os"
	"strings"
)

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

func GetEnvVarValue(key, dfault string) (envs string) {
	env := os.Environ()
	for _, envVar := range env {
		if strings.HasPrefix(envVar, key+"=") {
			return strings.Replace(envVar, key+"=", "", -1)
		}
	}
	return dfault
}
