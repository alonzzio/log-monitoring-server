package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGoRoutineID(t *testing.T) {
	got := GetGoRoutineID()
	assert.NotZero(t, got)
}
