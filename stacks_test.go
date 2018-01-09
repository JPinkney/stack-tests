package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/minishift/minishift/test/integration/util"
)

func tableRowGenerator(cellData WorkspaceTableItem) *gherkin.TableRow {

	var newTableCellNode gherkin.Node
	newTableCellNode.Type = "TableCell"

	var newCell gherkin.TableCell
	newCell.Node = newTableCellNode
	newCell.Value = cellData.Stack

	var newTableCellNode2 gherkin.Node
	newTableCellNode2.Type = "TableCell"

	var newCell2 gherkin.TableCell
	newCell2.Node = newTableCellNode
	newCell2.Value = cellData.ProjectName

	var newTableCellNode3 gherkin.Node
	newTableCellNode3.Type = "TableCell"

	var newCell3 gherkin.TableCell
	newCell3.Node = newTableCellNode
	newCell3.Value = cellData.Cmd

	var cells []*gherkin.TableCell
	cells = append(cells, &newCell, &newCell2, &newCell3)

	var newRow gherkin.TableRow
	newRow.Node = gherkin.Node{Type: "TableRow"}
	newRow.Cells = cells[0:]

	return &newRow

}

func tableRowArrayGenerator(cellDataArray []WorkspaceTableItem) []*gherkin.TableRow {

	var tableRowArray []*gherkin.TableRow

	for _, tableItem := range cellDataArray {

		newTableRow := tableRowGenerator(tableItem)
		tableRowArray = append(tableRowArray, newTableRow)

	}

	return tableRowArray
}

func setupExamplesData(g *gherkin.Feature) {
	WorkspaceTableItemArray, stackConfigInfo := testAllStacks("")
	StackConfigMap = stackConfigInfo
	for _, scenario := range g.ScenarioDefinitions {
		row := scenario.(*gherkin.ScenarioOutline).Examples[0].TableBody
		newTableRow := tableRowArrayGenerator(WorkspaceTableItemArray)

		scenario.(*gherkin.ScenarioOutline).Examples[0].TableBody = newTableRow
	}
}

// Workspace for finding out the workspace status
type Workspace struct {
	ID      string              `json:"id"`
	Config  WorkspaceConfig     `json:"workspaceConfig"`
	Source  WorkspaceSourceType `json:"source"`
	Tags    []string            `json:"tags"`
	Command []Command           `json:"commands"`
}

// Workspace for finding out the workspace status
type Workspace2 struct {
	ID string `json:"id"`
}

type StackConfigInfo struct {
	Config               WorkspaceConfig
	Projects             []Project
	WorkspaceSampleArray []WorkspaceSample
}

type Project struct {
	Sample interface{}
}

type WorkspaceSample struct {
	Config     WorkspaceConfig
	ID         string
	Sample     interface{}
	Command    []Command
	SamplePath string
}

// Workspace for finding out the workspace status
type WorkspaceStacks struct {
	Namespace  string                   `json:"namespace"`
	Status     string                   `json:"status"`
	Config     WorkspaceConfig          `json:"config"`
	Temporary  bool                     `json:"temporary"`
	ID         string                   `json:"id"`
	Attributes map[string]interface{}   `json:"attributes"`
	Links      []map[string]interface{} `json:"links"`
}

type Sample struct {
	Name     string           `json:"name"`
	Source   SampleSourceType `json:"source"`
	Commands []Command        `json:"commands"`
	Tags     []string         `json:"tags"`
	Path     string           `json:"path"`
}

// WorkspaceConfig is a config
type WorkspaceConfig struct {
	EnvironmentConfig EnvironmentConfig   `json:"environments"`
	Name              string              `json:"name"`
	DefaultEnv        string              `json:"defaultEnv"`
	Description       interface{}         `json:"description"`
	Commands          []Command           `json:"commands"`
	Source            WorkspaceSourceType `json:"source"`
}

type Command struct {
	CommandLine string `json:"commandLine"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type WorkspaceSourceType struct {
	Type   string `json:"type"`
	Origin string `json:"origin"`
}

type SampleSourceType struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

var rhStackLocation = "https://raw.githubusercontent.com/redhat-developer/rh-che/master/assembly/fabric8-stacks/src/main/resources/stacks.json"
var eclipseStackLocation = "http://localhost:8080/api/stack"
var samples = "https://raw.githubusercontent.com/eclipse/che/master/ide/che-core-ide-templates/src/main/resources/samples.json"
var tableData runArgsData
var fullyQualifiedEndpoint = "http://localhost:8080/api"
var StackConfigMap map[string]StackConfigInfo

// getJSON gets the json from URL and returns it
// To use the JSON you need to UnMarshall the response into your object
func getJSON(url string) []byte {

	client := http.Client{
		Timeout: time.Second * 60,
	}

	buf2 := new(bytes.Buffer)
	req, err := http.NewRequest(http.MethodGet, url, buf2)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	return body

}

type RuntimeStruct struct {
	Runtime Machine `json:"runtime"`
}

type Machine struct {
	Machines map[string]Servers `json:"machines"`
}

type Servers struct {
	Servers map[string]ServerURL `json:"servers"`
}

type ServerURL struct {
	URL string `json:"url"`
}

type Agent struct {
	execAgentURL string
	wsAgentURL   string
}

func getExecAgentHTTP(workspaceID string) (Agent, error) {
	var agents Agent

	//Now we need to get the workspace installers and then unmarshall
	runtimeData := getJSON(fullyQualifiedEndpoint + "/workspace/" + workspaceID)

	//fmt.Printf(string(runtimeData))
	var data RuntimeStruct
	jsonErr := json.Unmarshal(runtimeData, &data)
	if jsonErr != nil {
		return agents, fmt.Errorf("Could not unmrshall data into RuntimeStruct: %v", jsonErr)
	}

	for key := range data.Runtime.Machines {
		//fmt.Printf("%v", machine)
		//fmt.Printf("%v", data.Runtime.Machines[key])
		for key2, installer := range data.Runtime.Machines[key].Servers {

			if key2 == "exec-agent/http" {
				agents.execAgentURL = installer.URL
			}

			if key2 == "wsagent/http" {
				agents.wsAgentURL = installer.URL
			}

		}
	}

	return agents, nil
}

type ProcessStruct struct {
	Pid         int    `json:"pid"`
	Name        string `json:"name"`
	CommandLine string `json:"commandLine"`
	Type        string `json:"type"`
	Alive       bool   `json:"alive"`
	NativePid   int    `json:"nativePid"`
	ExitCode    int    `json:"exitCode"`
}

func postCommandToWorkspace(workspaceID, execAgentURL string, sampleCommand Command, samplePath string) int {

	//fmt.Printf(execAgentURL)
	sampleCommand.CommandLine = strings.Replace(sampleCommand.CommandLine, "${current.project.path}", "/projects"+samplePath, -1)
	sampleCommand.CommandLine = strings.Replace(sampleCommand.CommandLine, "${GAE}", "/home/user/google_appengine", -1)
	sampleCommand.CommandLine = strings.Replace(sampleCommand.CommandLine, "$TOMCAT_HOME", "/home/user/tomcat8", -1)
	marshalled, _ := json.MarshalIndent(sampleCommand, "", "    ")
	req, err := http.NewRequest("POST", execAgentURL, bytes.NewBufferString(string(marshalled)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	var data ProcessStruct
	//fmt.Printf(string(body))
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		panic(err.Error())
	}

	defer resp.Body.Close()

	return data.Pid

}

func checkCommandExitCode(Pid int, execAgentURL string) ProcessStruct {
	jsonData := getJSON(execAgentURL + "/" + strconv.Itoa(Pid))
	var data ProcessStruct
	jsonErr := json.Unmarshal(jsonData, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	if data.ExitCode > 0 {
		checkExecStatus(Pid, data.ExitCode, execAgentURL)
	}

	return data

}

func (stackRuntimeInfo *stackTestRuntimeInfo) continuouslyCheckCommandExitCode(Pid int, execAgentURL string) error {
	runCommand := checkCommandExitCode(Pid, execAgentURL)
	time.Sleep(15 * time.Second)
	checkExecStatus(Pid, runCommand.ExitCode, execAgentURL)
	for runCommand.Alive == true {
		time.Sleep(15 * time.Second)
		runCommand = checkCommandExitCode(Pid, execAgentURL)
		checkExecStatus(Pid, runCommand.ExitCode, execAgentURL)
	}

	stackRuntimeInfo.CommandExitCode = runCommand.ExitCode

	return nil
}

type LogArray []struct {
	Kind int       `json:"kind"`
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

func checkExecStatus(Pid, status int, execAgentURL string) {
	if status > 0 {
		jsonData := getJSON(execAgentURL + "/" + strconv.Itoa(Pid) + "/logs")
		var data LogArray
		jsonErr := json.Unmarshal(jsonData, &data)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}

		var buffer bytes.Buffer
		for _, value := range data {
			buffer.WriteString(value.Text)
			buffer.WriteString("\n")
		}

		fmt.Printf(buffer.String())
	}

}

func getSamplesJSON(url string) []Sample {

	client := http.Client{
		Timeout: time.Second * 60,
	}

	buf2 := new(bytes.Buffer)
	req, err := http.NewRequest(http.MethodGet, url, buf2)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	var data []Sample
	jsonErr := json.Unmarshal([]byte(body), &data)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data

}

type WorkspaceTableItem struct {
	Stack       string
	ProjectName string
	Cmd         string
}

func generateExampleTables(stackData []Workspace, samples []Sample, tag string) ([]WorkspaceTableItem, map[string]StackConfigInfo) {
	var tableElements []WorkspaceTableItem
	var stackConfigInfo map[string]StackConfigInfo
	for _, stackElement := range stackData {
		var samplesForStack []Project
		var workspaceSampleElements []WorkspaceSample
		for _, sampleElement := range samples {
			if len(sampleElement.Tags) != 0 {
				shouldAdd := false //Just in case two tags inside the same stack/sample element combo are the same

				//Finding whether atleast one of the tags match between the two
				for _, stackTag := range stackElement.Tags {
					for _, sampleTag := range sampleElement.Tags {
						if !shouldAdd && (strings.ToLower(stackTag) == strings.ToLower(sampleTag) || (stackTag != "" && (strings.ToLower(stackTag) == tag || strings.ToLower(sampleTag) == tag))) {
							shouldAdd = true
						}
					}
				}

				if shouldAdd {
					availableCommands := append(sampleElement.Commands, stackElement.Command...)

					//Prepend the build becaues the project has to be built before using other commands
					commandList := orderCommands(availableCommands)
					for _, cmd := range commandList {

						tableElements = append(tableElements, WorkspaceTableItem{
							Stack:       stackElement.ID,
							ProjectName: sampleElement.Path,
							Cmd:         cmd.CommandLine,
						})

					}

					workspaceSampleElements = append(workspaceSampleElements, WorkspaceSample{
						Command:    commandList,
						Config:     stackElement.Config,
						ID:         stackElement.ID,
						Sample:     sampleElement,
						SamplePath: sampleElement.Path,
					})

				}
			} else {
				availableCommands := append(sampleElement.Commands, stackElement.Command...)

				//Prepend the build becaues the project has to be built before using other commands
				commandList := orderCommands(availableCommands)
				for _, cmd := range commandList {

					tableElements = append(tableElements, WorkspaceTableItem{
						Stack:       stackElement.ID,
						ProjectName: sampleElement.Path,
						Cmd:         cmd.CommandLine,
					})

				}

				workspaceSampleElements = append(workspaceSampleElements, WorkspaceSample{
					Command:    commandList,
					Config:     stackElement.Config,
					ID:         stackElement.ID,
					Sample:     sampleElement,
					SamplePath: sampleElement.Path,
				})
			}

			samplesForStack = append(samplesForStack, Project{})

		}
		stackConfigInfo[stackElement.ID] = StackConfigInfo{
			Config:               stackElement.Config,
			WorkspaceSampleArray: workspaceSampleElements,
		}
	}
	return tableElements, stackConfigInfo
}

func orderCommands(commands []Command) []Command {
	var orderedCommands []Command
	for _, command := range commands {
		if strings.Contains(command.Name, "build") {
			orderedCommands = append([]Command{command}, orderedCommands...)
		} else {
			orderedCommands = append(orderedCommands, command)
		}
	}
	return orderedCommands
}

func execWithPiping(runCommandArgsSplit []string) (string, error) {
	dockerExecOutput := exec.Command(runCommandArgsSplit[0], runCommandArgsSplit[1:]...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	dockerExecOutput.Stdout = &stdout
	dockerExecOutput.Stderr = &stderr
	err := dockerExecOutput.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s", err)
	}

	return stdout.String(), nil
}

// Post is a post
type Post struct {
	Environments interface{}   `json:"environments"`
	Namespace    string        `json:"namespace"`
	Name         string        `json:"name"`
	DefaultEnv   string        `json:"defaultEnv"`
	Projects     []interface{} `json:"projects"`
}

type Commands struct {
	Name        string `json:"name"`
	CommandLine string `json:"commandLine"`
	Type        string `json:"type"`
}

// EnvironmentConfig is a config
type EnvironmentConfig struct {
	Default map[string]interface{} `json:"default"`
}

// WorkspaceStatus helps unmarshal workspace IDs to check if a given workspace is running
type WorkspaceStatus struct {
	WorkspaceStatus string `json:"status"`
}

type stackTestRuntimeInfo struct {
	ID                        string
	ExecAgentURL              string
	WSAgentURL                string
	WorkspaceStartingID       string
	CommandExitCode           int
	WorkspaceStopStatusCode   int
	WorkspaceRemoveStatusCode int
}

func (stackRuntimeInfo *stackTestRuntimeInfo) triggerStackStart(workspaceConfiguration StackConfigInfo, sample interface{}) (Workspace2, error) {
	workspaceConfig := workspaceConfiguration.Config
	test, err1 := json.Marshal(workspaceConfig)
	if err1 != nil {
		log.Fatal(err1)
	}

	jsonBytes := []byte(string(test))
	WorkspaceConfigInterface := &WorkspaceConfig{}
	json.Unmarshal(jsonBytes, WorkspaceConfigInterface)

	a := Post{Environments: WorkspaceConfigInterface.EnvironmentConfig, Namespace: "che", Name: stackRuntimeInfo.ID + "-stack-test", DefaultEnv: "default"}
	marshalled, _ := json.MarshalIndent(a, "", "    ")
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")
	req, err := http.NewRequest("POST", fullyQualifiedEndpoint+"/workspace?start-after-create=true", bytes.NewBufferString(noBayesian))

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	defer resp.Body.Close()

	var WorkspaceResponse Workspace2
	json.Unmarshal(buf.Bytes(), &WorkspaceResponse)

	return WorkspaceResponse, nil
}

func testAllStacks(tag string) ([]WorkspaceTableItem, map[string]StackConfigInfo) {
	stackData := getJSON(eclipseStackLocation)
	var data []Workspace
	jsonErr := json.Unmarshal(stackData, &data)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	samples := getSamplesJSON(samples)
	return generateExampleTables(data, samples, tag)
}

func addSampleToProject(wsAgentURL string, sample []WorkspaceSample) error {
	marshalled, _ := json.MarshalIndent(sample, "", "    ")
	req, err := http.NewRequest("POST", wsAgentURL+"/project/batch", bytes.NewBufferString(string(marshalled)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not complete the http request: %v", err)
	}

	defer resp.Body.Close()

	return nil
}

func getWorkspaceStatusByID(workspaceID string) (WorkspaceStatus, error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	var data WorkspaceStatus

	buf2 := new(bytes.Buffer)
	url := fullyQualifiedEndpoint + "/workspace/" + workspaceID
	req, err := http.NewRequest(http.MethodGet, url, buf2)
	if err != nil {
		return data, fmt.Errorf("Could not retrieve contents at url: %s with error %v", url, err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		return data, fmt.Errorf("Could not retrieve contents at url: %s with error %v", url, getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return data, fmt.Errorf("Could not retrieve response body: %v", readErr)
	}

	jsonErr := json.Unmarshal([]byte(body), &data)

	if jsonErr != nil {
		return data, fmt.Errorf("Could not unmarshal contents into WorkspaceStatus: %v", jsonErr)
	}

	return data, nil
}

func blockWorkspaceUntilStarted(workspaceID string) error {
	workspaceStatus, err := getWorkspaceStatusByID(workspaceID)
	if err != nil {
		return err
	}
	for workspaceStatus.WorkspaceStatus == "STARTING" {
		time.Sleep(30 * time.Second)
		workspaceStatus, err = getWorkspaceStatusByID(workspaceID)
		if err != nil {
			return err
		}
	}
	return nil
}

func blockWorkspaceUntilStopped(workspaceID string) error {
	workspaceStatus, err := getWorkspaceStatusByID(workspaceID)
	if err != nil {
		return err
	}
	//Workspace hasn't quite shut down due to speed
	for workspaceStatus.WorkspaceStatus == "SNAPSHOTTING" {
		time.Sleep(15 * time.Second)
		workspaceStatus, err = getWorkspaceStatusByID(workspaceID)
		if err != nil {
			return err
		}
	}

	time.Sleep(15 * time.Second)

	if workspaceStatus.WorkspaceStatus != "STOPPED" {
		return fmt.Errorf("Workspace was not stopped")
	}
	return nil
}

func (m *Minishift) executingSucceeds(addonInstall string) error {
	return minishift.executingMinishiftCommand(addonInstall)
}

func stdoutShouldContain(addonInstallResp string) error {
	return commandReturnShouldContain(addonInstallResp, addonInstallResp)
}

func (m *Minishift) minishiftHasState(runningState string) error {
	return m.shouldHaveState(runningState)
}

func (m *Minishift) minishiftShouldHaveState(runningState string) error {
	return m.shouldHaveState(runningState)
}

func (stackRuntimeInfo *stackTestRuntimeInfo) startingAWorkspaceWithStackPathAndCommandSucceeds(stack, path, command string) error {

	var workspaceConfigInfo = StackConfigMap[stack] //This is going to get you back the workspace config
	workspaceStartingResp, stackStartErr := stackRuntimeInfo.triggerStackStart(workspaceConfigInfo, path)
	if stackStartErr != nil {
		return fmt.Errorf("Problem starting the workspace: %v", stackStartErr)
	}

	blockingWorkspaceErr := blockWorkspaceUntilStarted(workspaceStartingResp.ID)
	if blockingWorkspaceErr != nil {
		return fmt.Errorf("Problem blocking the workspace until started: %v", stackStartErr)
	}

	agents, err := getExecAgentHTTP(workspaceStartingResp.ID)

	if err != nil {
		return err
	}

	for agents.execAgentURL == "" || agents.wsAgentURL == "" {
		agents, err = getExecAgentHTTP(workspaceStartingResp.ID)
		if err != nil {
			return err
		}
	}

	addingSampleError := addSampleToProject(agents.wsAgentURL, workspaceConfigInfo.WorkspaceSampleArray)
	if addingSampleError != nil {
		return addingSampleError
	}

	return nil

}

func (stackRuntimeInfo *stackTestRuntimeInfo) userRunsCommandOnPath(command Command, path string) error {
	Pid := postCommandToWorkspace(stackRuntimeInfo.WorkspaceStartingID, stackRuntimeInfo.ExecAgentURL, command, path)
	stackRuntimeInfo.continuouslyCheckCommandExitCode(Pid, stackRuntimeInfo.ExecAgentURL)
	return nil
}

func (stackRuntimeInfo *stackTestRuntimeInfo) commandShouldBeRanSuccessfully() error {
	if stackRuntimeInfo.CommandExitCode > 0 {
		return fmt.Errorf("Command did not complete") //Still need to get the logs and print them
	}
	return fmt.Errorf("Command did not complete")
}

func (stackRuntimeInfo *stackTestRuntimeInfo) userStopsWorkspace() error {
	url := fullyQualifiedEndpoint + "/workspace/" + stackRuntimeInfo.WorkspaceStartingID + "/runtime"
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	resp.Body.Close()

	blockWorkspaceUntilStopped(stackRuntimeInfo.WorkspaceStartingID)

	stackRuntimeInfo.WorkspaceStopStatusCode = resp.StatusCode

	return nil
}

func (stackRuntimeInfo *stackTestRuntimeInfo) workspaceStopShouldBeSuccessful() error {
	if stackRuntimeInfo.WorkspaceStopStatusCode >= 300 || stackRuntimeInfo.WorkspaceStopStatusCode < 200 {
		return fmt.Errorf("Could not stop the workspace")
	}
	return nil
}

func (stackRuntimeInfo *stackTestRuntimeInfo) workspaceIsRemoved() error {
	url := fullyQualifiedEndpoint + "/workspace/" + stackRuntimeInfo.WorkspaceStartingID
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	stackRuntimeInfo.WorkspaceRemoveStatusCode = resp.StatusCode
	defer resp.Body.Close()

	return nil
}

func (stackRuntimeInfo *stackTestRuntimeInfo) workspaceRemovalShouldBeSuccessful() error {
	if stackRuntimeInfo.WorkspaceRemoveStatusCode >= 300 || stackRuntimeInfo.WorkspaceRemoveStatusCode < 200 {
		return fmt.Errorf("Could not stop the workspace")
	}
	return nil
}

func FeatureContext(s *godog.Suite) {

	runner := util.MinishiftRunner{
		CommandArgs: minishiftArgs,
		CommandPath: minishiftBinary}

	minishift = &Minishift{runner: runner}

	stackRuntimeInfo := &stackTestRuntimeInfo{}

	s.be

	s.Step(`^executing "([^"]*)" succeeds$`, minishift.executingSucceeds)
	s.Step(`^stdout should contain "([^"]*)"$`, stdoutShouldContain)
	s.Step(`^Minishift has state "([^"]*)"$`, minishift.minishiftHasState)
	s.Step(`^Minishift should have state "([^"]*)"$`, minishift.minishiftShouldHaveState)
	s.Step(`^starting a workspace with stack "([^"]*)" path "([^"]*)" and command "([^"]*)" succeeds$`, stackRuntimeInfo.startingAWorkspaceWithStackPathAndCommandSucceeds)
	s.Step(`^user runs commands$`, stackRuntimeInfo.userRunsCommandOnPath)
	s.Step(`^command should be ran successfully$`, stackRuntimeInfo.commandShouldBeRanSuccessfully)
	s.Step(`^user stops workspace$`, stackRuntimeInfo.userStopsWorkspace)
	s.Step(`^workspace stop should be successful$`, stackRuntimeInfo.workspaceStopShouldBeSuccessful)
	s.Step(`^workspace is removed$`, stackRuntimeInfo.workspaceIsRemoved)
	s.Step(`^workspace removal should be successful$`, stackRuntimeInfo.workspaceRemovalShouldBeSuccessful)
}
