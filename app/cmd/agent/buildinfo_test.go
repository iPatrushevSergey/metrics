package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNA(t *testing.T) {
	assert.Equal(t, "N/A", buildNA(""))
	assert.Equal(t, "dev", buildNA("dev"))
}
