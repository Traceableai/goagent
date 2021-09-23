//go:build !linux
// +build !linux

package main

import "errors"

func getLinuxDistro() (string, error) {
	return "", errors.New("unsupported OS")
}
