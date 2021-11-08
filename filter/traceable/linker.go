//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func pathExists(p string) (bool, error) {
	if _, err := os.Stat(p); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func resolveLibPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// First we look for the library in the same directory as the executable.
	// This is mainly happening in run time.
	lookupPath := filepath.Join(filepath.Dir(execPath), "./libtraceable.so")
	ok, err := pathExists(lookupPath)
	if ok {
		return lookupPath, nil
	} else if err != nil {
		return "", err
	}

	// If library isn't present in the executable's directory, we look for
	// it in the package directory. This happens in build time and tests
	// where the directory for loading dynamic libraries is ${SRCDIR}.
	//
	// > The cgo tool will always invoke the C compiler with the source file's
	// directory in the include path; i.e. -I${SRCDIR} is always implied.
	//
	// See https://pkg.go.dev/cmd/cgo#hdr-Using_cgo_with_the_go_command
	useAlpine, err := isAlpine()
	if err != nil {
		return "", err
	}

	if useAlpine {
		lookupPath = "./libs/linux_amd64-alpine/libtraceable.so"
		ok, err = pathExists(lookupPath)
		if ok {
			return lookupPath, nil
		} else if err != nil {
			return "", err
		}
	} else {
		lookupPath = "./libs/linux_amd64/libtraceable.so"
		ok, err = pathExists(lookupPath)
		if ok {
			return lookupPath, nil
		} else if err != nil {
			return "", err
		}
	}

	return "", errors.New("libtraceable.so not found")
}

func isAlpine() (bool, error) {
	content, err := ioutil.ReadFile("/etc/os-release")
	if err == nil {
		return strings.Contains(string(content), "Alpine"), nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}
