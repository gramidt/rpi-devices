package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	_, err := getIP()
	assert.Error(t, err)
}
