//go:build !linux
// +build !linux

package main

import "errors"

func getLinuxDistroAndInstallDir() (string, string, error) {
	return "", "", errors.New("unsupported OS")
}
