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
)

type File struct {
  Name string
  Dir string
}

type Config struct {
  ShowTimer bool
  ShowScripts bool
}

func GetConfig() Config {
  cfg := Config{ true, true }
  home, err := os.UserHomeDir();
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

func GetAllProjects(dir string, level int ) []File {
  
  files, err := os.ReadDir(dir)
  if err != nil {
      log.Fatal(err)
  }

  projects := []File{}

  if IsProject(dir) {
      projects = append(projects, File{ path.Base(dir), dir })
  }

  for _, file := range files {
    if ! file.IsDir() {
      continue;
    }

    projectDir := path.Join(dir, file.Name())

    if ! IsProject(projectDir) {
      if level < 3 && !slices.Contains(BLACKLIST, file.Name()) {
        projects = append(projects, GetAllProjects(projectDir, level + 1)...)
      }
      continue;
    }


    projects = append(projects, File{ file.Name(), projectDir })
  }

  return projects;
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

