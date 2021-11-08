//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable // import "github.com/Traceableai/goagent/filter/traceable"

import (
	"errors"
	"io/ioutil"
	"os"
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
	lookupPath := "./libtraceable.so"
	ok, err := pathExists(lookupPath)
	if ok {
		return lookupPath, nil
	} else if err != nil {
		return "", err
	}

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
