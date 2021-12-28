package pst

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateServicesPool(t *testing.T) {
	r := Repo
	got := r.GenerateServicesPool(2)
	assert.Len(t, *got, 2)
}

func TestSeverityToString(t *testing.T) {
	r := Repo
	var tests = []struct {
		severity       Severity
		expectedOutput string
	}{
		{severity: 0, expectedOutput: "Debug"},
		{severity: 1, expectedOutput: "Info"},
		{severity: 2, expectedOutput: "Warn"},
		{severity: 3, expectedOutput: "Error"},
		{severity: 4, expectedOutput: "Fatal"},
	}

	for _, tt := range tests {
		actualOutput := r.SeverityToString(tt.severity)
		// Make sure our output matches
		if actualOutput != tt.expectedOutput {
			t.Errorf("Should have gotten expected output")
		}
	}
}
