package db

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewConn(t *testing.T) {
	err1 := os.Setenv("MYSQLROOTPASS", "example")
	err2 := os.Setenv("MYSQLDBNAME", "lms")
	_, err3 := NewConn()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
}
