package gira

import (
	"path/filepath"
	"strings"
)

func findMatchingProject(projects []ProjectConfig, cwd, homeDir string) *ProjectConfig {
	normalizedCwd := normalizePath(cwd)

	var best *ProjectConfig
	bestLength := 0

	for index := range projects {
		length := projectMatchLength(projects[index], normalizedCwd, homeDir)
		if length > bestLength {
			best, bestLength = &projects[index], length
		}
	}

	return best
}

func projectMatchLength(project ProjectConfig, cwd, homeDir string) int {
	candidates := []string{project.Path}
	if project.WorktreeBasePath != "" {
		candidates = append(candidates, project.WorktreeBasePath)
	}

	longest := 0
	for _, candidate := range candidates {
		candidate = normalizePath(expandHome(candidate, homeDir))
		if pathMatches(candidate, cwd) && len(candidate) > longest {
			longest = len(candidate)
		}
	}

	return longest
}

func expandHome(path, homeDir string) string {
	if path == "~" {
		return homeDir
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	}

	return path
}

func normalizePath(path string) string {
	if strings.Contains(path, `\`) {
		return path
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}

	return absolute
}

func pathMatches(projectPath, cwd string) bool {
	if projectPath == cwd {
		return true
	}

	separator := "/"
	if strings.Contains(projectPath, `\`) {
		separator = `\`
	}
	if !strings.HasSuffix(projectPath, separator) {
		projectPath += separator
	}

	return strings.HasPrefix(cwd, projectPath)
}
