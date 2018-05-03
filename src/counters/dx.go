package counters

import (
	"archive/zip"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	st "github.com/palantir/stacktrace"

	"github.com/dhleong/dexcounter/src/model"
	"github.com/dhleong/dexcounter/src/util"
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

	// easy case
	if strings.HasSuffix(dep.Path, ".jar") {
		return dc.checkJar(dep, dep.Path)
	}

	// handle .aar
	if strings.HasSuffix(dep.Path, ".aar") {
		dir, err := util.GetConfigDir("aars")
		if err != nil {
			return err
		}

		jarName := strings.Replace(dep.Dependency.String(), ":", "-", -1)
		jarPath := filepath.Join(dir, fmt.Sprintf("%s.jar", jarName))
		foundJar, err := extractClassesJarTo(dep.Path, jarPath)
		if err != nil {
			return err
		}
		if !foundJar {
			// resources-only .aar
			return nil
		}

		return dc.checkJar(dep, jarPath)
	}

	// something else
	return fmt.Errorf("Unknown dependency format: %s", dep.Path)
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

// Returns True if there were classes in the .aar
func extractClassesJarTo(aarPath, destPath string) (bool, error) {

	if _, err := os.Stat(destPath); err == nil {
		// jar already exists!
		return true, nil
	}

	zipFile, err := zip.OpenReader(aarPath)
	if err != nil {
		return false, err
	}
	defer zipFile.Close()

	for _, f := range zipFile.File {
		if f.Name == "classes.jar" {
			zipFp, err := f.Open()
			if err != nil {
				return false, err
			}
			defer zipFp.Close()

			destFp, err := os.Create(destPath)
			if err != nil {
				return false, err
			}
			defer destFp.Close()

			_, err = io.Copy(destFp, zipFp)
			if err != nil {
				return false, err
			}

			// done!
			return true, nil
		}
	}

	// no classes; probably resources-only .aar
	return false, nil
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
