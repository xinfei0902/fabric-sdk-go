package tools

import (
	"os"
	"path/filepath"
	"strings"
)

func STDPath(name string) string {
	if len(name) == 0 {
		return name
	}
	if false == filepath.IsAbs(name) {
		base, _ := filepath.Split(os.Args[0])
		name = filepath.Join(base, name)
	}
	return name
}

func STDString(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}
