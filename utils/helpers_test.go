package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func makeProject(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"package.json", "composer.json"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func projectNames(projects []File) []string {
	names := make([]string, 0, len(projects))
	for _, project := range projects {
		names = append(names, project.Name)
	}
	return names
}

func TestGetAllProjectsOnlyReturnsProjects(t *testing.T) {
	root := t.TempDir()
	makeProject(t, filepath.Join(root, "app"))
	if err := os.MkdirAll(filepath.Join(root, "not-a-project"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := projectNames(GetAllProjects(root, 1, 0, false))
	want := []string{"app"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAllProjects() = %v, want %v", got, want)
	}
}

func TestGetAllProjectsHonorsDepth(t *testing.T) {
	root := t.TempDir()
	makeProject(t, filepath.Join(root, "one", "two", "app"))

	if got := GetAllProjects(root, 2, 0, false); len(got) != 0 {
		t.Fatalf("depth 2 returned projects at depth 3: %v", projectNames(got))
	}
	if got := projectNames(GetAllProjects(root, 3, 0, false)); !reflect.DeepEqual(got, []string{"app"}) {
		t.Fatalf("depth 3 returned %v, want [app]", got)
	}
	if got := projectNames(GetAllProjects(root, -1, 0, false)); !reflect.DeepEqual(got, []string{"app"}) {
		t.Fatalf("unlimited depth returned %v, want [app]", got)
	}
}

func TestGetAllProjectsExcludesCurrentAndBlacklistedTrees(t *testing.T) {
	root := t.TempDir()
	makeProject(t, root)
	makeProject(t, filepath.Join(root, "child"))
	makeProject(t, filepath.Join(root, "node_modules", "dependency"))

	got := projectNames(GetAllProjects(root, -1, 0, true))
	want := []string{"child"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAllProjects() = %v, want %v", got, want)
	}
}
