package gitssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

const (
	envVarName = "GIT_SSH"
	safeMode   = 0600
	execMode   = 0700
)

// Wrapper is an object capable of temporarily changing the environment to allow
// using a string SSH key to use with Git.
type Wrapper interface {
	// Cleanup removes all files used by the Wrapper. Using it invalidates
	// the Wrapper.
	Cleanup() error

	// Environment returns an environment with just the GIT_SSH variable
	// updated. If passed a nil argument it will use os.Environ() as base,
	// otherwise it will use whatever is passed as an argument.
	Environment([]string) []string

	// GitSSH returns the value of the script which - if set - will have Git
	// use a non-standard SSH connection.
	GitSSH() string

	// Link exports the GIT_SSH environment variable to the current session.
	Link() error

	// Unlink resets the GIT_SSH environment variable in the current
	// session.
	Unlink() error
}

type tmplArgs struct {
	HostsFile, KeyFile string
}

type wrapperImpl struct {
	hostsFilePath, privateKeyPath, scriptPath, oldGitSSH string
}

func tempfile(prefix, content string, mode os.FileMode) (string, error) {
	file, err := ioutil.TempFile("", prefix)
	if err != nil {
		return "", err
	}
	if _, err := file.Write([]byte(content)); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	path := file.Name()
	if err := os.Chmod(path, mode); err != nil {
		return "", err
	}
	return path, nil
}

// NewWrapper uses a private key to create a new SSH wrapper.
func NewWrapper(privateKey string) (Wrapper, error) {
	hostsFilePath, err := tempfile("hosts", "", safeMode)
	if err != nil {
		return nil, err
	}
	privateKeyPath, err := tempfile("key", privateKey, safeMode)
	if err != nil {
		return nil, err
	}
	wrapper := &wrapperImpl{
		hostsFilePath:  hostsFilePath,
		privateKeyPath: privateKeyPath,
	}
	if err := wrapper.createScript(); err != nil {
		return nil, err
	}
	return wrapper, nil
}

func (w *wrapperImpl) Cleanup() error {
	for _, f := range []string{w.hostsFilePath, w.privateKeyPath, w.scriptPath} {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func (w *wrapperImpl) Environment(original []string) []string {
	if original == nil {
		original = os.Environ()
	}
	var ret []string
	needle := fmt.Sprintf("%s=", envVarName)
	for _, pair := range original {
		if !strings.HasPrefix(pair, needle) {
			ret = append(ret, pair)
		}
	}
	return append(ret, fmt.Sprintf("%s%s", needle, w.GitSSH()))
}

func (w *wrapperImpl) GitSSH() string {
	return w.scriptPath
}

func (w *wrapperImpl) Link() error {
	w.oldGitSSH = os.Getenv(envVarName)
	return os.Setenv(envVarName, w.scriptPath)
}

func (w *wrapperImpl) Unlink() error {
	return os.Setenv(envVarName, w.oldGitSSH)
}

func (w *wrapperImpl) templateArgs() *tmplArgs {
	return &tmplArgs{w.hostsFilePath, w.privateKeyPath}
}

func (w *wrapperImpl) createScript() error {
	tmpl, err := template.ParseFiles("ssh.template")
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, w.templateArgs()); err != nil {
		return err
	}
	w.scriptPath, err = tempfile("script", buffer.String(), execMode)
	return err
}

// ExecuteWithSSKey runs a callback function (presumably a Git command) with a
// given SSH key.
func ExecuteWithSSKey(privateKey string, fn func()) error {
	wrapper, err := NewWrapper(privateKey)
	if err != nil {
		return err
	}
	if err := wrapper.Link(); err != nil {
		return err
	}
	defer wrapper.Unlink()
	defer wrapper.Cleanup()
	fn()
	return nil
}
