package counters

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/dhleong/dexcounter/src/model"
)

type gradleDexCounter struct {
	workspaceDir string
}

// NewGradleDexCounter usestes a DexCounter that uses gradle
// to determine the dependencies of a given library. Note that
// this means that ALL returned Count instances will have a value
// of 0 for any dex counts!
func NewGradleDexCounter() (model.DexCounter, error) {
	dir, err := ensureGradleSetUp()
	if err != nil {
		return nil, err
	}

	return gradleDexCounter{dir}, nil
}

func (dc gradleDexCounter) Count(dep model.Dependency) (*model.Counts, error) {
	gradlew := filepath.Join(dc.workspaceDir, "gradlew")
	cmd := exec.Command(
		gradlew,
		"-p", dc.workspaceDir,
		"-q",
		"deps",
		fmt.Sprintf("-PinputDep=%s", dep),
	)
	output, err := cmd.Output()
	if err != nil {
		if v, ok := err.(*exec.ExitError); ok {
			fmt.Println(string(v.Stderr))
		}
		return nil, err
	}
	return parseOutput(output)
}

func parseOutput(output []byte) (*model.Counts, error) {
	reader := bufio.NewReader(bytes.NewReader(output))

	var root *model.Counts

	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		if err == io.EOF || line == "" {
			break
		} else if err != nil {
			return nil, err
		}

		if root == nil {
			root = parseLine(line)
		} else {
			root.Dependents = append(root.Dependents, parseLine(line))
		}
	}

	return root, nil
}

func parseLine(line string) *model.Counts {
	parts := strings.Split(line, "|")
	return &model.Counts{
		Dependency: model.Dependency{
			Group:    parts[0],
			Artifact: parts[1],
			Version:  parts[2],
		},
		Path: parts[3],
	}
}

func ensureGradleSetUp() (string, error) {

	// first, check if we're running locally,
	// for development purposes
	if _, err := os.Stat("gradle"); err == nil {
		// aha!
		return "gradle", nil
	}

	// now, fall back to real implementation, downloading the
	// gradle files from github to a local directory
	dir, err := homedir.Expand("~/.config/dexcounter/gradle")
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

	// TODO download

	return "", errors.New("Not implemented yet")
}
