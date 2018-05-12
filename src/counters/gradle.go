package counters

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dhleong/dexcounter/src/model"
	"github.com/dhleong/dexcounter/src/util"
)

// baseURL links to a specific revision to ensure compatibility
const commitHash = "e73d86ddd404adaf3daec9fcd1947f20c2cc5c5f"
const baseURL = "https://raw.githubusercontent.com/dhleong/dexcounter/" + commitHash + "/gradle/"

// these are suffixes to baseURL
var gradleFileURLs = []string{
	"gradlew.bat",
	"gradlew",
	"gradle.properties",
	"build.gradle",
}

var gradleWrapperFileURLs = []string{
	"gradle/wrapper/gradle-wrapper.properties",
	"gradle/wrapper/gradle-wrapper.jar",
}

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

func (dc gradleDexCounter) Count(
	dep model.Dependency,
	ui model.UI,
) (*model.Counts, error) {
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
		return "gradle", nil
	}

	// normal mode, okay create the dirs
	gradleFilesDir, err := util.GetConfigDir("gradle/" + commitHash)
	if err != nil {
		return "", err
	}

	gradleWrapperDir, err := util.GetConfigDir(
		"gradle/" + commitHash + "/gradle/wrapper",
	)
	if err != nil {
		return "", err
	}

	var allDone sync.WaitGroup
	allDone.Add(len(gradleFileURLs) + len(gradleWrapperFileURLs))

	// fetch all the files in parallel
	var someErr error
	for _, url := range gradleFileURLs {
		go func(url string) {
			err := downloadTo(url, gradleFilesDir)
			if err != nil {
				someErr = err
			}
			allDone.Done()
		}(url)
	}

	for _, url := range gradleWrapperFileURLs {
		go func(url string) {
			downloadTo(url, gradleWrapperDir)
			allDone.Done()
		}(url)
	}

	allDone.Wait()

	// make them executable
	if err := os.Chmod(
		filepath.Join(gradleFilesDir, "gradlew"),
		0774,
	); err != nil {
		return "", err
	}

	if err := os.Chmod(
		filepath.Join(gradleFilesDir, "gradlew.bat"),
		0774,
	); err != nil {
		return "", err
	}

	return gradleFilesDir, someErr
}

func downloadTo(url, destDir string) error {

	name := path.Base(url)
	destFile := path.Join(destDir, name)

	// create the local file
	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// request the remote file
	var fullURL = baseURL + url
	resp, err := http.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// write the response body
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
