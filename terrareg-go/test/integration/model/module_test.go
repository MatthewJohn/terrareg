package model

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/stretchr/testify/assert"
)

// TestModule_InvalidNames tests that invalid module names are rejected
func TestModule_InvalidNames(t *testing.T) {
	invalidNames := []string{
		"invalid@atsymbol",
		"invalid\"doublequote",
		"invalid'singlequote",
		"-startwithdash",
		"endwithdash-",
		"_startwithunderscore",
		"endwithunscore_",
		"a:colon",
		"or;semicolon",
		"who?knows",
		"-a",
		"a-",
		"a_",
		"_a",
		"__",
		"--",
		"_",
		"-",
	}

	for _, name := range invalidNames {
		name := name // capture range variable
		t.Run(name, func(t *testing.T) {
			t.Parallel() // Domain validation tests can run in parallel
			err := model.ValidateModuleName(name)
			assert.Error(t, err, "Expected error for invalid module name: %s", name)
		})
	}
}

// TestModule_ValidNames tests that valid module names are accepted
func TestModule_ValidNames(t *testing.T) {
	validNames := []string{
		"normalname",
		"name2withnumber",
		"2startendiwthnumber2",
		"contains4number",
		"with-dash",
		"with_underscore",
		"withAcapital",
		"StartwithCaptital",
		"endwithcapitaL",
		"tl",      // Two letters
		"11",      // Two numbers
		"a-z",     // Two characters with dash
		"a_z",     // Two characters with underscore
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := model.ValidateModuleName(name)
			assert.NoError(t, err, "Expected no error for valid module name: %s", name)
		})
	}
}
