package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadURL(t *testing.T) {
	expectedURL := "https://traceableai.jfrog.io/artifactory/gradle-local/ai/traceable/agent/libtraceable_ubuntu_20.04_x86_64/0.1.96/libtraceable_ubuntu_20.04_x86_64-0.1.96.zip"
	actualURL := downloadURL(libraryInfo{
		Name:    "libtraceable",
		OS:      "ubuntu_20.04",
		Version: "0.1.96",
	})
	assert.Equal(t, expectedURL, actualURL)
}

func TestUnzip(t *testing.T) {
	dirpath, err := ioutil.TempDir(os.TempDir(), "")
	require.NoError(t, err)
	c, err := unzipFile("./testdata/test.zip", "test.txt", dirpath)
	assert.NoError(t, err)
	assert.Equal(t, 0, c)
	os.RemoveAll(dirpath)

	dirpath, err = ioutil.TempDir(os.TempDir(), "")
	require.NoError(t, err)
	c, err = unzipFile("./testdata/test.zip", "test/test1.txt", dirpath)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
	os.RemoveAll(dirpath)

	dirpath, err = ioutil.TempDir(os.TempDir(), "")
	require.NoError(t, err)
	c, err = unzipFile("./testdata/test.zip", "test/sub/test2.txt", dirpath)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
	os.RemoveAll(dirpath)
}

func TestOsDetection(t *testing.T) {
	os, _ := getOSAndDestinationDir()
	t.Fail(os)
}
