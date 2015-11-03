package gitssh

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const sshKey = "haxxpr00f"

func TestBasicLifecycle(t *testing.T) {
	wrapper, err := NewWrapper(sshKey)
	if err != nil {
		t.Fatalf("NewWrapper err = %v, expected nil", err)
	}
	gitSSH := wrapper.GitSSH()
	fi, err := os.Stat(gitSSH)
	if err != nil {
		t.Errorf("os.Stat(gitSSH) err = %v, expected nil", err)
	}
	if fi.Mode() != execMode {
		t.Errorf("fi.Mode = %d, expected %d", fi.Mode(), execMode)
	}
	bytes, err := ioutil.ReadFile(gitSSH)
	if err != nil {
		t.Fatalf("ioutil.ReadFile(gitSSH) err = %v, expected nil", err)
	}
	if !strings.HasPrefix(string(bytes), "#!/bin/sh") {
		t.Errorf("unexpected script content: %s", bytes)
	}
	if err := wrapper.Cleanup(); err != nil {
		t.Fatalf("wrapper.Cleanup err = %v, expected nil", err)
	}
}

func TestLinkAndUnlink(t *testing.T) {
	wrapper, _ := NewWrapper(sshKey)
	defer wrapper.Cleanup()
	oldKey := os.Getenv(envVarName)
	if err := wrapper.Link(); err != nil {
		t.Fatalf("wrapper.Link() = %v, expected nil", err)
	}
	if os.Getenv(envVarName) == oldKey {
		t.Errorf("expected environment variable %q to change", envVarName)
	}
	if err := wrapper.Unlink(); err != nil {
		t.Fatalf("wrapper.Unlink() = %v, expected nil", err)
	}
}

func TestExecutionWrapper(t *testing.T) {
	oldKey := os.Getenv(envVarName)
	var script string
	err := ExecuteWithSSKey(sshKey, func() {
		script = os.Getenv(envVarName)
		if script == "" {
			t.Errorf("environment variable %q not set", envVarName)
		}
	})
	if err != nil {
		t.Fatalf("ExecuteWithSSKey = %v, expected nil", err)
	}
	newKey := os.Getenv(envVarName)
	if oldKey != newKey {
		t.Errorf("environment variable %q changed from %q to %q", envVarName, oldKey, newKey)
	}
}
