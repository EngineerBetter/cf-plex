package target

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Add(plexHome, api string) (string, error) {
	apiDir := Sanitise(api)
	fullPath := filepath.Join(plexHome, apiDir)
	err := os.MkdirAll(fullPath, 0700)
	return fullPath, err
}

func Remove(plexHome, api string) error {
	apiDir := Sanitise(api)
	fullPath := filepath.Join(plexHome, apiDir)
	return os.RemoveAll(fullPath)
}

func Sanitise(apiUrl string) string {
	api := strings.Replace(apiUrl, "https://", "https___", -1)
	api = strings.Replace(api, "http://", "http___", -1)
	return api
}

func List(plexHome string) ([]string, error) {
	f, err := os.Open(plexHome)
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
		names[index] = filepath.Join(plexHome, apiDir)
	}

	return names, nil
}
