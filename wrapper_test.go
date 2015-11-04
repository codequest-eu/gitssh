package gitssh

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const sshKey = "haxxpr00f"

type WrapperTestSuite struct {
	suite.Suite
	wrapper Wrapper
}

func (s *WrapperTestSuite) SetupTest() {
	wrapper, err := NewWrapper(sshKey)
	s.Require().Nil(err)
	s.wrapper = wrapper
}

func (s *WrapperTestSuite) TearDownTest() {
	s.Nil(s.wrapper.Cleanup())
}

func (s *WrapperTestSuite) TestGitSSH() {
	gitSSH := s.wrapper.GitSSH()
	fi, err := os.Stat(gitSSH)
	s.Require().Nil(err)
	s.EqualValues(execMode, fi.Mode())
	content, err := ioutil.ReadFile(gitSSH)
	s.Require().Nil(err)
	s.Contains(string(content), "#!/bin/sh")
}

func (s *WrapperTestSuite) TestEnvironmentAdd() {
	newEnv := s.wrapper.Environment([]string{})
	s.Len(newEnv, 1)
	s.Contains(newEnv, fmt.Sprintf("GIT_SSH=%s", s.wrapper.GitSSH()))
}

func (s *WrapperTestSuite) TestEnvironmentReplace() {
	newEnv := s.wrapper.Environment([]string{"GIT_SSH=bacon"})
	s.Len(newEnv, 1)
	s.Contains(newEnv, fmt.Sprintf("GIT_SSH=%s", s.wrapper.GitSSH()))
}

func (s *WrapperTestSuite) TestLinkAndUnlink() {
	oldKey := os.Getenv(envVarName)
	s.Require().Nil(s.wrapper.Link())
	s.NotEqual(oldKey, os.Getenv(envVarName))
	s.Require().Nil(s.wrapper.Unlink())
	s.Equal(oldKey, os.Getenv(envVarName))
}

func TestExecutionWrapper(t *testing.T) {
	oldKey := os.Getenv(envVarName)
	var scriptPath string
	err := ExecuteWithSSKey(sshKey, func() {
		scriptPath = os.Getenv(envVarName)
		assert.NotEqual(t, "", scriptPath)
	})
	require.Nil(t, err)
	assert.Equal(t, oldKey, os.Getenv(envVarName))
}

func TestWrapper(t *testing.T) {
	suite.Run(t, new(WrapperTestSuite))
}
