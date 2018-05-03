package util

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
)

// GetConfigDir gets the path to a subdir of our ".config" directory,
// creating it and any parents if necessary
func GetConfigDir(dirName string) (string, error) {
	// now, fall back to real implementation, downloading the
	// gradle files from github to a local directory
	dir, err := homedir.Expand(fmt.Sprintf("~/.config/dexcounter/%s", dirName))
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			// some unexpected issue; report it:
			return "", err
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
	}

	return dir, nil
}
