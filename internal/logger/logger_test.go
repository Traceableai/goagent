package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	assert.Nil(t, _l)
	GetLogger()
	assert.NotNil(t, _l)
}
