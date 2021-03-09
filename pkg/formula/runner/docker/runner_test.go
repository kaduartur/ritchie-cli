/*
 * Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package docker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ZupIT/ritchie-cli/pkg/api"
	"github.com/ZupIT/ritchie-cli/pkg/env"
	"github.com/ZupIT/ritchie-cli/pkg/formula"
	"github.com/ZupIT/ritchie-cli/pkg/formula/builder"
	"github.com/ZupIT/ritchie-cli/pkg/formula/input/flag"
	"github.com/ZupIT/ritchie-cli/pkg/formula/input/prompt"
	"github.com/ZupIT/ritchie-cli/pkg/formula/input/stdin"
	"github.com/ZupIT/ritchie-cli/pkg/formula/runner"
	"github.com/ZupIT/ritchie-cli/pkg/stream"
	"github.com/ZupIT/ritchie-cli/pkg/stream/streams"
)

func TestRun(t *testing.T) {
	fileManager := stream.NewFileManager()
	dirManager := stream.NewDirManager(fileManager)
	tmpDir := os.TempDir()
	homeDir, _ := os.UserHomeDir()
	ritHome := filepath.Join(tmpDir, ".rit-runner-docker")
	repoPath := filepath.Join(ritHome, "repos", "commons")

	dockerBuilder := builder.NewBuildDocker(fileManager)

	_ = dirManager.Remove(ritHome)
	_ = dirManager.Remove(repoPath)
	_ = dirManager.Create(repoPath)
	zipFile := filepath.Join("..", "..", "..", "..", "testdata", "ritchie-formulas-test.zip")
	_ = streams.Unzip(zipFile, repoPath)

	envFinder := env.NewFinder(ritHome, fileManager)
	preRunner := NewPreRun(ritHome, dockerBuilder, dirManager, fileManager)

	pInputRunner := prompt.NewInputManager(envResolverMock{in: "test"}, inputMock{}, inputMock{}, inputTextValidatorMock{str: "test"}, inputTextDefaultMock{}, inputMock{}, inputMock{}, inputMock{})
	sInputRunner := stdin.NewInputManager(envResolverMock{in: "test"})
	fInputRunner := flag.NewInputManager(envResolverMock{in: "test"})

	types := formula.TermInputTypes{
		api.Prompt: pInputRunner,
		api.Stdin:  sInputRunner,
		api.Flag:   fInputRunner,
	}
	inputResolver := runner.NewInputResolver(types)

	type in struct {
		def           formula.Definition
		preRun        formula.PreRunner
		inputResolver formula.InputResolver
		env           env.Finder
	}

	type out struct {
		err error
	}

	tests := []struct {
		name string
		in   in
		out  out
	}{
		{
			name: "run docker success",
			in: in{
				def:           formula.Definition{Path: "testing/formula", RepoName: "commons"},
				preRun:        preRunner,
				inputResolver: inputResolver,
				env:           envFinder,
			},
			out: out{
				err: nil,
			},
		},
		{
			name: "input error",
			in: in{
				def:           formula.Definition{Path: "testing/formula", RepoName: "commons"},
				preRun:        preRunner,
				inputResolver: inputResolverMock{err: runner.ErrInputNotRecognized},
				env:           envFinder,
			},
			out: out{
				err: runner.ErrInputNotRecognized,
			},
		},
		{
			name: "pre run error",
			in: in{
				def:           formula.Definition{Path: "testing/formula", RepoName: "commons"},
				preRun:        preRunnerMock{err: errors.New("pre runner error")},
				inputResolver: inputResolver,
				env:           envFinder,
			},
			out: out{
				err: errors.New("pre runner error"),
			},
		},
		{
			name: "env find error",
			in: in{
				def:           formula.Definition{Path: "testing/formula", RepoName: "commons"},
				preRun:        preRunner,
				inputResolver: inputResolver,
				env:           envFinderMock{err: errors.New("env not found")},
			},
			out: out{
				err: errors.New("env not found"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwd, _ := os.Getwd()
			formulaPath := tt.in.def.FormulaPath(ritHome)
			_ = os.Chdir(formulaPath)

			in := tt.in
			docker := NewRunner(homeDir, in.inputResolver, in.preRun, in.env)
			got := docker.Run(in.def, api.Prompt, false, nil)

			if got != nil || tt.out.err != nil {
				assert.EqualError(t, tt.out.err, got.Error())
			} else {
				fileCreated := filepath.Join(formulaPath, "test.txt")
				assert.FileExists(t, fileCreated)
				_ = os.Remove(fileCreated)
			}

			_ = os.Chdir(pwd)
		})
	}

}

type preRunnerMock struct {
	setup formula.Setup
	err   error
}

func (pr preRunnerMock) PreRun(def formula.Definition) (formula.Setup, error) {
	return pr.setup, pr.err
}

type envResolverMock struct {
	in  string
	err error
}

func (e envResolverMock) Resolve(string) (string, error) {
	return e.in, e.err
}

type inputTextValidatorMock struct {
	str string
}

func (i inputTextValidatorMock) Text(name string, validate func(interface{}) error, helper ...string) (string, error) {
	return i.str, nil
}

type inputMock struct {
	text    string
	boolean bool
	items   []string
	err     error
}

func (i inputMock) List(string, []string, ...string) (string, error) {
	return i.text, i.err
}

func (i inputMock) Text(string, bool, ...string) (string, error) {
	return i.text, i.err
}

func (i inputMock) Bool(string, []string, ...string) (bool, error) {
	return i.boolean, i.err
}

func (i inputMock) Password(string, ...string) (string, error) {
	return i.text, i.err
}

func (i inputMock) Multiselect(formula.Input) ([]string, error) {
	return i.items, i.err
}

type inputTextDefaultMock struct {
	text string
	err  error
}

func (i inputTextDefaultMock) Text(formula.Input) (string, error) {
	return i.text, i.err
}

type envFinderMock struct {
	env env.Holder
	err error
}

func (c envFinderMock) Find() (env.Holder, error) {
	return c.env, c.err
}

type inputResolverMock struct {
	inRunner formula.InputRunner
	err      error
}

func (i inputResolverMock) Resolve(inType api.TermInputType) (formula.InputRunner, error) {
	return i.inRunner, i.err
}
