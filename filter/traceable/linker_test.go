//go:build linux && traceable_filter
// +build linux,traceable_filter

package traceable

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPathExists(t *testing.T) {
	f, err := ioutil.TempFile("", "tmpfile-")
	require.NoError(t, err)

	found, err := pathExists(f.Name())
	assert.NoError(t, err)
	assert.True(t, found)

	f.Close()

	os.Remove(f.Name())
	found, err = pathExists(f.Name())
	assert.NoError(t, err)
	assert.False(t, found)
}
