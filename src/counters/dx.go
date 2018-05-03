package counters

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	st "github.com/palantir/stacktrace"

	"github.com/dhleong/dexcounter/src/model"
)

// DxDexCounter uses the dx tool, provided with the Android SDK,
// to check dex counts
type DxDexCounter struct {
	depsSource model.DexCounter
	dxPath     string
}

// NewDxDexCounter .
func NewDxDexCounter(
	opts *model.Options,
	dependenciesSource model.DexCounter,
) (model.DexCounter, error) {

	dx, err := pickDxPath(opts)
	if err != nil {
		return nil, err
	}

	return &DxDexCounter{
		dependenciesSource,
		dx,
	}, nil
}

// Count .
func (dc *DxDexCounter) Count(
	dependency model.Dependency,
) (*model.Counts, error) {

	dep, err := dc.depsSource.Count(dependency)
	if err != nil {
		return nil, err
	}

	// count all dependencies in parallel
	flattened := dep.Flatten()
	totalCount := len(flattened)

	var allDone sync.WaitGroup
	allDone.Add(totalCount)

	var someErr error
	for _, counts := range flattened {
		go func(counts *model.Counts) {
			err := dc.count(counts)
			if err != nil {
				someErr = st.Propagate(err, "Error checking %v", counts.Dependency)
			}
			allDone.Done()
		}(counts)
	}

	allDone.Wait()

	return dep, someErr
}

func (dc *DxDexCounter) count(dep *model.Counts) error {

	if dep.Path == "" {
		return fmt.Errorf("No Path for %v", dep.Dependency)
	}

	if strings.HasSuffix(dep.Path, ".jar") {
		return dc.checkJar(dep, dep.Path)
	}

	// FIXME TODO handle .aar

	return nil
}

func (dc *DxDexCounter) checkJar(dep *model.Counts, jarPath string) error {
	cmd := exec.Command(dc.dxPath, "--dex", "--output=-", jarPath)
	bytes, err := cmd.Output()
	if err != nil {
		// see: https://github.com/dextorer/MethodsCount/blob/master/app/services/sdk_service.rb
		// some libraries require --core-library
		cmd = exec.Command(dc.dxPath, "--dex", "--output=-", jarPath)
		bytes, err = cmd.Output()
		if err != nil {
			// log stderr:
			if v, ok := err.(*exec.ExitError); ok {
				fmt.Println("Running", cmd)
				fmt.Println(string(v.Stderr))
			}
			return err
		}
	}

	// see dex format: https://source.android.com/devices/tech/dalvik/dex-format#header-item
	fieldsBytes := bytes[80:84]
	methodsBytes := bytes[88:92]
	dep.OwnFields = int(binary.LittleEndian.Uint32(fieldsBytes))
	dep.OwnMethods = int(binary.LittleEndian.Uint32(methodsBytes))

	return nil
}

func pickDxPath(opts *model.Options) (string, error) {

	if opts.DxPath != "" {
		// verify it exists
		if _, err := os.Stat(opts.DxPath); err == nil {
			// cool!
			return opts.DxPath, nil
		}

		return "", fmt.Errorf(
			"Provided dx path is invalid: %s",
			opts.DxPath,
		)
	}

	// try $ANDROID_HOME
	androidHome := os.Getenv("ANDROID_HOME")
	if androidHome != "" {
		if fromHome := pickDxPathFromHome(androidHome); fromHome != "" {
			return fromHome, nil
		}
	}

	return "", errors.New("Unable to locate `dx`")
}

func pickDxPathFromHome(home string) string {
	matches, err := filepath.Glob(fmt.Sprintf("%s/build-tools/*/dx", home))
	if err != nil || len(matches) == 0 {
		// no matches
		return ""
	}

	return matches[len(matches)-1]
}
