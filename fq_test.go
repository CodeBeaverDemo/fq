package main

import (
    "testing"
    "github.com/wader/fq/pkg/cli"
    "github.com/wader/fq/pkg/interp"
)

// TestMainCallsCliMain tests that main() calls cliMain with the correct parameters.
func TestMainCallsCliMain(t *testing.T) {
var cliMain = cli.Main

    var called bool
    mainFunc := func() {
        cliMain(interp.DefaultRegistry, version)
    }
    origCliMain := cliMain
    defer func() { cliMain = origCliMain }()

    cliMain = func(reg *interp.Registry, ver string) {
    called = true
    if reg != interp.DefaultRegistry {
    t.Errorf("Expected interp.DefaultRegistry, got %v", reg)
    }
    if ver != version {
    t.Errorf("Expected version %s, got %s", version, ver)
    }
    }

    mainFunc()

    if !called {
    t.Error("cliMain was not called by main()")
    }
}

// TestVersion tests that the version constant is set to the expected value.
func TestVersion(t *testing.T) {
    expected := "0.12.0"
    if version != expected {
    t.Errorf("Expected version %s, got %s", expected, version)
    }
}