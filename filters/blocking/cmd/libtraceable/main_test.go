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
	c, err := unzip("./testdata/test.zip", "", dirpath)
	assert.NoError(t, err)
	assert.Equal(t, 2, c)
	os.RemoveAll(dirpath)

	dirpath, err = ioutil.TempDir(os.TempDir(), "")
	require.NoError(t, err)
	c, err = unzip("./testdata/test.zip", "test", dirpath)
	assert.NoError(t, err)
	assert.Equal(t, 2, c)
	os.RemoveAll(dirpath)

	dirpath, err = ioutil.TempDir(os.TempDir(), "")
	require.NoError(t, err)
	c, err = unzip("./testdata/test.zip", "test/sub", dirpath)
	assert.NoError(t, err)
	assert.Equal(t, 1, c)
	os.RemoveAll(dirpath)
}
