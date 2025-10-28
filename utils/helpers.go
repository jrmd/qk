/*
Copyright © 2025 Jerome Duncan <jerome@jrmd.dev>
*/
package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path"
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

	if ok, err := FileExists(path.Join(home, ".qk.json")); !ok || err != nil {
		return cfg
	}

	conf, err := os.ReadFile(path.Join(home, ".qk.json"))

	if err != nil {
		return cfg
	}

	_ = json.Unmarshal(conf, &cfg)
	return cfg
}

var BLACKLIST = []string{"node_modules", ".git", ".idea", "vendor"}

func GetAllProjects(dir string, depth int, level int) []File {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	projects := []File{}

	if IsProject(dir) {
		projects = append(projects, File{path.Base(dir), dir})
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		projectDir := path.Join(dir, file.Name())

		if !IsProject(projectDir) && ( depth == -1 || level <= depth ) {
			if !slices.Contains(BLACKLIST, file.Name()) {
				projects = append(projects, GetAllProjects(projectDir, depth, level + 1)...)
			}
			continue
		}

		if depth != -1 && level >= depth {
			continue
		}

		projects = append(projects, File{file.Name(), projectDir})
	}

	return projects
}

func IsProject(dir string) bool {
	hasComposer, _ := FileExists(path.Join(dir, "composer.json"))
	hasPackage, _ := FileExists(path.Join(dir, "package.json"))
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
	for _, t := range ts {
		if pred(t) {
			return true
		}
	}
	return false
}

func HasYarn(project types.Project) bool {
	exists, _ := FileExists(path.Join(project.Dir, "yarn.lock"))
	return exists
}

func Not[T any](pred func(T) bool) func(T) bool {
	return func(thing T) bool {
		return !pred(thing)
	}
}

func And[T any](preds ...func(T) bool) func(T) bool {
	return func(thing T) bool {
		return All(preds, func (pred func(T) bool) bool {
			return pred(thing)
		})
	}
}

func HasScript(script string) func(p types.Project) bool {
	return func (project types.Project) bool {
		file, err := os.ReadFile(path.Join(project.Dir, "package.json"))
		if err != nil {
			return false
		}
		pkg := PackageJSON{}
		_ = json.Unmarshal(file, &pkg)
		_, exists := pkg.Scripts[script]

		return exists
	}
}
