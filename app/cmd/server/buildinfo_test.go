package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNA(t *testing.T) {
	assert.Equal(t, "N/A", buildNA(""))
	assert.Equal(t, "1.0", buildNA("1.0"))
}

func TestPrintBuildInfo(t *testing.T) {
	printBuildInfo()
}
