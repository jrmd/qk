/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package utils

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"jrmd.dev/qk/types"
)

type File struct {
	Name string
	Dir  string
}

type Config struct {
	ShowTimer   bool
	ShowScripts bool
	ShowStdout  bool
}

type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

func GetConfig() Config {
	cfg := Config{true, true, false}
	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	if ok, err := FileExists(filepath.Join(home, ".qk.json")); !ok || err != nil {
		return cfg
	}

	conf, err := os.ReadFile(filepath.Join(home, ".qk.json"))

	if err != nil {
		return cfg
	}

	_ = json.Unmarshal(conf, &cfg)
	return cfg
}

var BLACKLIST = []string{"node_modules", ".git", ".idea", "vendor"}

func GetAllProjects(dir string, depth int, level int, excludeCurrent bool) []File {
	if depth < -1 || level < 0 {
		return nil
	}

	projects := []File{}
	if IsProject(dir) && !excludeCurrent {
		projects = append(projects, File{filepath.Base(filepath.Clean(dir)), dir})
	}

	// A depth of zero searches only dir. A depth of -1 is unlimited.
	if depth != -1 && level >= depth {
		return projects
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return projects
	}

	for _, file := range files {
		if !file.IsDir() || slices.Contains(BLACKLIST, file.Name()) {
			continue
		}

		projectDir := filepath.Join(dir, file.Name())
		if IsProject(projectDir) {
			projects = append(projects, File{file.Name(), projectDir})
			continue
		}

		projects = append(projects, GetAllProjects(projectDir, depth, level+1, false)...)
	}

	return projects
}

func IsProject(dir string) bool {
	hasComposer, _ := FileExists(filepath.Join(dir, "composer.json"))
	hasPackage, _ := FileExists(filepath.Join(dir, "package.json"))
	return hasComposer && hasPackage
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func All[T any](ts []T, pred func(T) bool) bool {
	for _, t := range ts {
		if !pred(t) {
			return false
		}
	}
	return true
}

func Some[T any](ts []T, pred func(T) bool) bool {
	return slices.ContainsFunc(ts, pred)
}

func HasYarn(project types.Project) bool {
	exists, _ := FileExists(filepath.Join(project.Dir, "yarn.lock"))
	return exists
}

func Not[T any](pred func(T) bool) func(T) bool {
	return func(thing T) bool {
		return !pred(thing)
	}
}

func And[T any](preds ...func(T) bool) func(T) bool {
	return func(thing T) bool {
		return All(preds, func(pred func(T) bool) bool {
			return pred(thing)
		})
	}
}

func HasScript(script string) func(p types.Project) bool {
	return func(project types.Project) bool {
		file, err := os.ReadFile(filepath.Join(project.Dir, "package.json"))
		if err != nil {
			return false
		}
		pkg := PackageJSON{}
		_ = json.Unmarshal(file, &pkg)
		_, exists := pkg.Scripts[script]

		return exists
	}
}
