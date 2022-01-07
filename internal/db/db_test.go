package db

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewConn(t *testing.T) {
	os.Setenv("MYSQLROOTPASS", "example")
	os.Setenv("MYSQLDBNAME", "lms")
	_, err := NewConn()
	assert.NoError(t, err)
}
