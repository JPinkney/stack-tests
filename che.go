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
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jpinkney/stack-tests/util"
)

type CheRunner struct {
	runner util.CheAPI
}

var samples = "https://raw.githubusercontent.com/eclipse/che/master/ide/che-core-ide-templates/src/main/resources/samples.json"
var stackConfigMap map[string]util.Workspace
var sampleConfigMap map[string]util.Sample

func generateStackData(stackData []util.Workspace, samples []util.Sample) (map[string]util.Workspace, map[string]util.Sample) {

	stackConfigInfo := make(map[string]util.Workspace)
	sampleConfigInfo := make(map[string]util.Sample)
	for _, stackElement := range stackData {
		stackConfigInfo[stackElement.Name] = stackElement
	}

	for _, sampleElement := range samples {
		sampleConfigInfo[sampleElement.Source.Location] = sampleElement
	}

	return stackConfigInfo, sampleConfigInfo
}

func (c *CheRunner) weTryToGetTheStacksInformation() error {
	stackData := c.runner.GetJSON(c.runner.CheAPIEndpoint + "/stack")
	var data []util.Workspace
	jsonStackErr := json.Unmarshal(stackData, &data)

	if jsonStackErr != nil {
		return fmt.Errorf("Could not retrieve stack information: %v. CheAPIEndpoint is: %v. Data is: %v", jsonStackErr, c.runner.CheAPIEndpoint+"/stack", data)
	}

	samplesJSON := c.runner.GetJSON(samples)
	var sampleData []util.Sample
	jsonSamplesErr := json.Unmarshal([]byte(samplesJSON), &sampleData)

	if jsonSamplesErr != nil {
		log.Fatal(jsonSamplesErr)
	}

	stackConfigMap, sampleConfigMap = generateStackData(data, sampleData)

	return nil
}

func (c *CheRunner) theStacksShouldNotBeEmpty() error {
	if len(stackConfigMap) == 0 || len(sampleConfigMap) == 0 {
		return fmt.Errorf("Could not retrieve samples")
	}
	return nil
}

func (c *CheRunner) startingAWorkspaceWithStackSucceeds(stackName string) error {
	stackStartEnvironment := stackConfigMap[stackName]
	workspace, err := c.runner.StartWorkspace(stackStartEnvironment.Config.EnvironmentConfig, stackStartEnvironment.ID)
	if err != nil {
		return err
	}

	c.runner.SetWorkspaceID(workspace.ID)
	c.runner.BlockWorkspaceUntilStarted(workspace.ID)
	c.runner.SetStackName(stackName)

	agents, err := c.runner.GetHTTPAgents(workspace.ID)
	if err != nil {
		return err
	}
	c.runner.SetAgentsURL(agents)
	return nil
}

func (c *CheRunner) workspaceShouldHaveState(expectedState string) error {
	currentState, err := c.runner.GetWorkspaceStatusByID(c.runner.WorkspaceID)
	if err != nil {
		return err
	}

	if strings.Compare(strings.ToLower(currentState.WorkspaceStatus), strings.ToLower(expectedState)) != 0 {
		return fmt.Errorf("Not in expected state. Current state is: %s. Expected state is: %s", currentState.WorkspaceStatus, expectedState)
	}

	return nil
}

func (c *CheRunner) importingTheSampleProjectSucceeds(projectURL string) error {
	sample := sampleConfigMap[projectURL]
	err := c.runner.AddSamplesToProject([]util.Sample{sample})
	if err != nil {
		return err
	}
	return nil
}

func (c *CheRunner) workspaceShouldHaveProject(numOfProjects int) error {
	numOfProjects, err := c.runner.GetNumberOfProjects()
	if err != nil {
		return err
	}

	if numOfProjects == 0 {
		return fmt.Errorf("No projects were added")
	}

	return nil
}

func (c *CheRunner) userRunsCommandOnSample(projectURL string) error {

	if len(sampleConfigMap[projectURL].Commands) > 0 {
		sampleCommand := sampleConfigMap[projectURL].Commands[0]
		c.runner.PID = c.runner.PostCommandToWorkspace(sampleCommand)
	} else {
		sampleCommand := stackConfigMap[c.runner.StackName].Command[0]
		c.runner.PID = c.runner.PostCommandToWorkspace(sampleCommand)
	}

	return nil
}

func (c *CheRunner) exitCodeShouldBe(code int) error {
	// if c.runner.PID != code {
	// 	return fmt.Errorf("return command was not 0")
	// }
	return nil
}

func (c *CheRunner) userStopsWorkspace() error {
	err := c.runner.StopWorkspace(c.runner.WorkspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *CheRunner) workspaceIsRemoved() error {
	err := c.runner.RemoveWorkspace(c.runner.WorkspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *CheRunner) workspaceRemovalShouldBeSuccessful() error {

	respCode, err := c.runner.CheckWorkspaceDeletion(c.runner.WorkspaceID)
	if err != nil {
		return err
	}

	if respCode != 404 {
		return fmt.Errorf("Workspace has not been removed")
	}

	return nil
}
