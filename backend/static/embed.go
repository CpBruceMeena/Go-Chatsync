package static

import (
	"embed"
	"io/fs"
)

//go:embed build/*
var buildFS embed.FS

// GetBuildFS returns the embedded filesystem for the React build directory
func GetBuildFS() (fs.FS, error) {
	return fs.Sub(buildFS, "build")
}
