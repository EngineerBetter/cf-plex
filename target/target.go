package target

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Target struct {
	Name string
	Path string
}

type Group struct {
	Name string
	Apis []Target
}

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

func List(plexHome string) ([]Group, error) {
	var groups []Group

	if GroupsExist(plexHome) {
		dirs, err := listDirs(filepath.Join(plexHome, "groups"))
		if err != nil {
			return nil, err
		}

		for _, groupDir := range dirs {
			targets, err := getTargets(groupDir)
			if err != nil {
				return nil, err
			}

			groupName := filepath.Base(groupDir)
			if groupIsPublic(groupName) {
				groups = append(groups, Group{Name: filepath.Base(groupDir), Apis: targets})
			}
		}
	} else {
		targets, err := getTargets(plexHome)
		if err != nil {
			return nil, err
		}
		groups = append(groups, Group{Name: "default", Apis: targets})
	}

	return groups, nil
}

func GroupsExist(plexHome string) bool {
	groupsPath := filepath.Join(plexHome, "groups")
	_, err := os.Stat(groupsPath)

	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func AddToGroup(plexHome, group, api string) (string, error) {
	apiDir := Sanitise(api)
	fullPath := filepath.Join(plexHome, "groups", group, apiDir)
	err := os.MkdirAll(fullPath, 0700)
	return fullPath, err
}

func RemoveFromGroup(plexHome, group, api string) error {
	apiDir := Sanitise(api)
	fullPath := filepath.Join(plexHome, "groups", group, apiDir)
	err := os.RemoveAll(fullPath)
	if err != nil {
		return err
	}

	groupDir := filepath.Join(plexHome, "groups", group)
	dirs, err := listDirs(groupDir)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		err = os.RemoveAll(groupDir)
	}
	return err
}

func Sanitise(apiUrl string) string {
	api := strings.Replace(apiUrl, "https://", "https___", -1)
	api = strings.Replace(api, "http://", "http___", -1)
	return api
}

func MakeFilthy(apiDir string) string {
	api := strings.Replace(apiDir, "https___", "https://", -1)
	api = strings.Replace(api, "http___", "http://", -1)
	return api
}

func getTargets(parentPath string) ([]Target, error) {
	apiDirs, err := listDirs(parentPath)
	if err != nil {
		return nil, err
	}

	var targets []Target
	for _, apiDir := range apiDirs {
		name := MakeFilthy(path.Base(apiDir))
		targets = append(targets, Target{Name: name, Path: apiDir})
	}
	return targets, nil
}

func groupIsPublic(groupName string) bool {
	return groupName != "batch"
}

func listDirs(path string) ([]string, error) {
	f, err := os.Open(path)
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
		names[index] = filepath.Join(path, apiDir)
	}

	return names, nil
}
