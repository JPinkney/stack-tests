package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
)

var samples = "https://raw.githubusercontent.com/eclipse/che/master/ide/che-core-ide-templates/src/main/resources/samples.json"
var stackConfigMap map[string]Workspace
var sampleConfigMap map[string]Sample

func (c *cheRunner) getStackInformation(arg interface{}) {
	stackData := getJSON(c.cheAPIEndpoint + "/stack")
	var data []Workspace
	jsonStackErr := json.Unmarshal(stackData, &data)

	if jsonStackErr != nil {
		log.Fatal(jsonStackErr)
	}

	samplesJSON := getJSON(samples)
	var sampleData []Sample
	jsonSamplesErr := json.Unmarshal([]byte(samplesJSON), &sampleData)

	if jsonSamplesErr != nil {
		log.Fatal(jsonSamplesErr)
	}

	stackConfigMap, sampleConfigMap = generateStackData(data, sampleData)
}

func generateStackData(stackData []Workspace, samples []Sample) (map[string]Workspace, map[string]Sample) {

	stackConfigInfo := make(map[string]Workspace)
	sampleConfigInfo := make(map[string]Sample)
	for _, stackElement := range stackData {
		stackConfigInfo[stackElement.Name] = stackElement
	}

	for _, sampleElement := range samples {
		sampleConfigInfo[sampleElement.Source.Location] = sampleElement
	}

	return stackConfigInfo, sampleConfigInfo
}

func executingSucceeds(command string) error {
	return nil
}

func stdoutShouldContain(commandReturn string) error {
	return nil
}

func minishiftHasState(state string) error {
	return nil
}

func minishiftShouldHaveState(state string) error {
	return nil
}

func (c *cheRunner) startingAWorkspaceWithStackSucceeds(stackName string) error {
	stackStartEnvironment := stackConfigMap[stackName]
	workspace, err := c.startWorkspace(stackStartEnvironment.Config.EnvironmentConfig, stackStartEnvironment.ID)
	if err != nil {
		return err
	}
	c.setWorkspaceID(workspace.ID)
	c.blockWorkspaceUntilStarted(workspace.ID)
	agents, err := c.getHTTPAgents(workspace.ID)
	if err != nil {
		return err
	}
	c.setAgentsURL(agents)
	return nil
}

func (c *cheRunner) workspaceShouldHaveState(expectedState string) error {
	currentState, err := c.getWorkspaceStatusByID(c.workspaceID)
	if err != nil {
		return err
	}

	if strings.Compare(strings.ToLower(currentState.WorkspaceStatus), strings.ToLower(expectedState)) != 0 {
		return fmt.Errorf("Not in expected state. Current state is: %s. Expected state is: %s", currentState.WorkspaceStatus, expectedState)
	}

	return nil
}

func (c *cheRunner) importingTheSampleProjectSucceeds(projectURL string) error {
	sample := sampleConfigMap[projectURL]
	err := c.addSamplesToProject([]Sample{sample})
	if err != nil {
		return err
	}
	return nil
}

func (c *cheRunner) workspaceShouldHaveProject(numOfProjects int) error {
	numOfProjects, err := c.getNumberOfProjects()
	if err != nil {
		return err
	}

	if numOfProjects == 0 {
		return fmt.Errorf("No projects were added")
	}

	return nil
}

func (c *cheRunner) userRunsCommand(projectURL string) error {
	sampleCommand := sampleConfigMap[projectURL].Commands[0]
	c.PID = c.postCommandToWorkspace(sampleCommand)
	return nil
}

func (c *cheRunner) exitCodeShouldBe(code int) error {
	if c.PID != code {
		return fmt.Errorf("return command was not 0")
	}
	return nil
}

func (c *cheRunner) userStopsWorkspace() error {
	err := c.stopWorkspace(c.workspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *cheRunner) workspaceIsRemoved() error {
	err := c.removeWorkspace(c.workspaceID)
	if err != nil {
		return err
	}
	return nil
}

func (c *cheRunner) workspaceRemovalShouldBeSuccessful() error {

	respCode, err := c.checkWorkspaceDeletion(c.workspaceID)
	if err != nil {
		return err
	}

	if respCode != 404 {
		return fmt.Errorf("Workspace has not been removed")
	}

	return nil
}

func FeatureContext(s *godog.Suite) {

	cheAPIRunner := &cheRunner{
		cheAPIEndpoint: "http://che-eclipse-che.192.168.42.24.nip.io/api",
	}

	//Minishift Tests
	s.Step(`^executing "([^"]*)" succeeds$`, executingSucceeds)
	s.Step(`^stdout should contain "([^"]*)"$`, stdoutShouldContain)
	s.Step(`^Minishift has state "([^"]*)"$`, minishiftHasState)
	s.Step(`^Minishift should have state "([^"]*)"$`, minishiftShouldHaveState)

	//Start of Che tests
	s.BeforeScenario(cheAPIRunner.getStackInformation)
	s.Step(`^starting a workspace with stack "([^"]*)" succeeds$`, cheAPIRunner.startingAWorkspaceWithStackSucceeds)
	s.Step(`^workspace should have state "([^"]*)"$`, cheAPIRunner.workspaceShouldHaveState)
	s.Step(`^importing the sample project "([^"]*)" succeeds$`, cheAPIRunner.importingTheSampleProjectSucceeds)
	s.Step(`^workspace should have (\d+) project$`, cheAPIRunner.workspaceShouldHaveProject)
	s.Step(`^user runs command$`, cheAPIRunner.userRunsCommand)
	s.Step(`^exit code should be (\d+)$`, cheAPIRunner.exitCodeShouldBe)
	s.Step(`^user stops workspace$`, cheAPIRunner.userStopsWorkspace)
	s.Step(`^workspace is removed$`, cheAPIRunner.workspaceIsRemoved)
	s.Step(`^workspace removal should be successful$`, cheAPIRunner.workspaceRemovalShouldBeSuccessful)
}

func TestMain(m *testing.M) {

	status := godog.RunWithOptions("godog", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format: "progress",
		Paths:  []string{"features"},
	})

	start := time.Now()
	if st := m.Run(); st > status {
		status = st
	}
	elapsed := time.Since(start)
	os.Exit(status)
	fmt.Printf("go test -all took %s", elapsed)
}
