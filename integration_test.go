/*
Copyright (C) 2017 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"
	"testing"

	"github.com/DATA-DOG/godog"

	"github.com/jpinkney/stack-tests/util"
)

func TestMain(m *testing.M) {

	status := godog.RunWithOptions("godog", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format: "progress",
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {

	// steps for testing che addon
	cheAPI := util.CheAPI{
		CheAPIEndpoint: "http://localhost:8081/api",
	}

	cheAPIRunner := &CheRunner{
		runner: cheAPI,
	}

	s.Step(`^we try to get the stacks information$`, cheAPIRunner.weTryToGetTheStacksInformation)
	s.Step(`^the stacks should not be empty$`, cheAPIRunner.theStacksShouldNotBeEmpty)
	s.Step(`^starting a workspace with stack "([^"]*)" succeeds$`, cheAPIRunner.startingAWorkspaceWithStackSucceeds)
	s.Step(`^workspace should have state "([^"]*)"$`, cheAPIRunner.workspaceShouldHaveState)
	s.Step(`^importing the sample project "([^"]*)" succeeds$`, cheAPIRunner.importingTheSampleProjectSucceeds)
	s.Step(`^workspace should have (\d+) project$`, cheAPIRunner.workspaceShouldHaveProject)
	s.Step(`^user runs command on sample "([^"]*)"$`, cheAPIRunner.userRunsCommandOnSample)
	s.Step(`^exit code should be (\d+)$`, cheAPIRunner.exitCodeShouldBe)
	s.Step(`^user stops workspace$`, cheAPIRunner.userStopsWorkspace)
	s.Step(`^workspace is removed$`, cheAPIRunner.workspaceIsRemoved)
	s.Step(`^workspace removal should be successful$`, cheAPIRunner.workspaceRemovalShouldBeSuccessful)

}
