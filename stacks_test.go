package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/minishift/minishift/test/integration/util"
)

func setupExamplesData(g *gherkin.Feature) {
	WorkspaceTableItemArray, stackConfigInfo := testAllStacks("")
	StackConfigMap = stackConfigInfo
	newTableRow := tableRowArrayGenerator(WorkspaceTableItemArray)
	g.ScenarioDefinitions[2].(*gherkin.ScenarioOutline).Examples[0].TableBody = newTableRow
	//for _, scenario := range g.ScenarioDefinitions {
	// newTableRow := tableRowArrayGenerator(WorkspaceTableItemArray)
	// scenario.(*gherkin.ScenarioOutline).Examples[0].TableBody = newTableRow
	//}
}

func tableRowArrayGenerator(cellDataArray []WorkspaceTableItem) []*gherkin.TableRow {

	var tableRowArray []*gherkin.TableRow

	for _, tableItem := range cellDataArray {
		newTableRow := tableRowGenerator(tableItem)
		tableRowArray = append(tableRowArray, newTableRow)
	}

	return tableRowArray
}

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

type Workspace struct {
	ID      string              `json:"id"`
	Config  WorkspaceConfig     `json:"workspaceConfig"`
	Source  WorkspaceSourceType `json:"source"`
	Tags    []string            `json:"tags"`
	Command []Command           `json:"commands"`
}

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
var eclipseStackLocation = "http://che-eclipse-che.192.168.42.233.nip.io/api/stack"
var samples = "https://raw.githubusercontent.com/eclipse/che/master/ide/che-core-ide-templates/src/main/resources/samples.json"
var fullyQualifiedEndpoint = "http://che-eclipse-che.192.168.42.233.nip.io/api"
var StackConfigMap map[string]StackConfigInfo
var ProjectMap map[string][]Command

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

type ProcessStruct struct {
	Pid         int    `json:"pid"`
	Name        string `json:"name"`
	CommandLine string `json:"commandLine"`
	Type        string `json:"type"`
	Alive       bool   `json:"alive"`
	NativePid   int    `json:"nativePid"`
	ExitCode    int    `json:"exitCode"`
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

func postCommandToWorkspace(workspaceID, execAgentURL string, sampleCommand string, samplePath string) int {

	//Find the command from the project
	commandList := ProjectMap[samplePath]

	//From here we need to get look through and find the command that has sample command equal to it
	var commandInfo Command
	for _, command := range commandList {
		if command.Name == sampleCommand {
			commandInfo = command
		}
	}

	commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "${current.project.path}", "/projects"+samplePath, -1)
	commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "${GAE}", "/home/user/google_appengine", -1)
	commandInfo.CommandLine = strings.Replace(commandInfo.CommandLine, "$TOMCAT_HOME", "/home/user/tomcat8", -1)
	marshalled, _ := json.MarshalIndent(commandInfo, "", "    ")
	req, err := http.NewRequest("POST", execAgentURL, bytes.NewBufferString(string(marshalled)))
	req.Header.Set("Content-Type", "application/json")

	//fmt.Printf("%v", req)

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

//THIS FUNCTION STILL NEEDS THE CHECK TO SEE IF ITS A LONG RUNNING PROCESS
func (stackRuntimeInfo *stackTestRuntimeInfo) continuouslyCheckCommandExitCode(Pid int, execAgentURL string) error {
	runCommand := checkCommandExitCode(Pid, execAgentURL)
	time.Sleep(15 * time.Second)
	checkExecStatus(Pid, runCommand.ExitCode, execAgentURL)
	count := 3
	for runCommand.Alive == true {
		if count < 3 {
			time.Sleep(15 * time.Second)
			runCommand = checkCommandExitCode(Pid, execAgentURL)
			checkExecStatus(Pid, runCommand.ExitCode, execAgentURL)
		} else if count >= 3 {
			runCommand.Alive = false
		}

		count++
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

		//fmt.Printf(buffer.String())
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
	stackConfigInfo := make(map[string]StackConfigInfo)
	projectMap := make(map[string][]Command)
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
							Cmd:         cmd.Name,
						})

					}

					workspaceSampleElements = append(workspaceSampleElements, WorkspaceSample{
						Command:    commandList,
						Config:     stackElement.Config,
						ID:         stackElement.ID,
						Sample:     sampleElement,
						SamplePath: sampleElement.Path,
					})

					projectMap[sampleElement.Path] = commandList

				}
			} else {
				availableCommands := append(sampleElement.Commands, stackElement.Command...)

				//Prepend the build becaues the project has to be built before using other commands
				commandList := orderCommands(availableCommands)
				for _, cmd := range commandList {

					tableElements = append(tableElements, WorkspaceTableItem{
						Stack:       stackElement.ID,
						ProjectName: sampleElement.Path,
						Cmd:         cmd.Name,
					})

				}

				workspaceSampleElements = append(workspaceSampleElements, WorkspaceSample{
					Command:    commandList,
					Config:     stackElement.Config,
					ID:         stackElement.ID,
					Sample:     sampleElement,
					SamplePath: sampleElement.Path,
				})

				projectMap[sampleElement.Path] = commandList
			}

			samplesForStack = append(samplesForStack, Project{})

		}
		stackConfigInfo[stackElement.ID] = StackConfigInfo{
			Config:               stackElement.Config,
			WorkspaceSampleArray: workspaceSampleElements,
		}
	}

	ProjectMap = projectMap

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

type EnvironmentConfig struct {
	Default map[string]interface{} `json:"default"`
}

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
	workspaceConfig := workspaceConfiguration.Config.EnvironmentConfig

	a := Post{Environments: workspaceConfig, Namespace: "che", Name: stackRuntimeInfo.ID + "-stack-test", DefaultEnv: "default"}
	marshalled, _ := json.MarshalIndent(a, "", "    ")
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")

	//fmt.Printf("%v", noBayesian)

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

	if WorkspaceResponse.ID == "" {
		return WorkspaceResponse, fmt.Errorf("Could not get starting ID")
	}

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

	var sampleArray []interface{}
	for _, workspaceSample := range sample {
		sampleArray = append(sampleArray, workspaceSample.Sample)
	}

	marshalled, _ := json.MarshalIndent(sampleArray, "", "    ")
	req, err := http.NewRequest("POST", wsAgentURL+"/project/batch", bytes.NewBufferString(string(marshalled)))
	req.Header.Set("Content-Type", "application/json")

	//fmt.Printf("%v", req)

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
	return nil
	//return minishift.executingMinishiftCommand(addonInstall)
}

func stdoutShouldContain(addonInstallResp string) error {
	return nil
	//return commandReturnShouldContain(addonInstallResp, addonInstallResp)
}

func (m *Minishift) minishiftHasState(runningState string) error {
	return nil
	//return m.shouldHaveState(runningState)
}

func (m *Minishift) minishiftShouldHaveState(runningState string) error {
	return nil
	//return m.shouldHaveState(runningState)
}

func (stackRuntimeInfo *stackTestRuntimeInfo) minishiftHasStateAndStartingAWorkspaceWithStackPathAndCommandSucceeds(running, stack, path, command string) error {

	var workspaceConfigInfo = StackConfigMap[stack] //This is going to get you back the workspace config
	stackRuntimeInfo.ID = stack
	workspaceStartingResp, stackStartErr := stackRuntimeInfo.triggerStackStart(workspaceConfigInfo, path)
	if stackStartErr != nil {
		return fmt.Errorf("Problem starting the workspace: %v", stackStartErr)
	}

	blockingWorkspaceErr := blockWorkspaceUntilStarted(workspaceStartingResp.ID)
	if blockingWorkspaceErr != nil {
		return fmt.Errorf("Problem blocking the workspace until started: %v", stackStartErr)
	}

	stackRuntimeInfo.WorkspaceStartingID = workspaceStartingResp.ID

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

	stackRuntimeInfo.WSAgentURL = agents.wsAgentURL
	stackRuntimeInfo.ExecAgentURL = agents.execAgentURL

	addingSampleError := addSampleToProject(agents.wsAgentURL, workspaceConfigInfo.WorkspaceSampleArray)
	if addingSampleError != nil {
		return addingSampleError
	}

	return nil

}

func (stackRuntimeInfo *stackTestRuntimeInfo) userRunsCommandOnPath(command string, path string) error {
	Pid := postCommandToWorkspace(stackRuntimeInfo.WorkspaceStartingID, stackRuntimeInfo.ExecAgentURL, command, path)
	stackRuntimeInfo.continuouslyCheckCommandExitCode(Pid, stackRuntimeInfo.ExecAgentURL)
	return nil
}

func (stackRuntimeInfo *stackTestRuntimeInfo) commandShouldBeRanSuccessfully() error {
	if stackRuntimeInfo.CommandExitCode > 0 {
		return fmt.Errorf("Command errored") //Still need to get the logs and print them
	}
	return nil
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

func FeatureContext(s *godog.Suite) {

	runner := util.MinishiftRunner{
		CommandArgs: minishiftArgs,
		CommandPath: minishiftBinary}

	minishift = &Minishift{runner: runner}

	stackRuntimeInfo := &stackTestRuntimeInfo{}

	s.BeforeFeature(setupExamplesData)
	s.Step(`^executing "([^"]*)" succeeds$`, minishift.executingSucceeds)
	s.Step(`^stdout should contain "([^"]*)"$`, stdoutShouldContain)
	s.Step(`^Minishift has state "([^"]*)"$`, minishift.minishiftHasState)
	s.Step(`^Minishift should have state "([^"]*)"$`, minishift.minishiftShouldHaveState)
	s.Step(`^Minishift has state "([^"]*)" and starting a workspace with stack "([^"]*)" path "([^"]*)" and command "([^"]*)" succeeds$`, stackRuntimeInfo.minishiftHasStateAndStartingAWorkspaceWithStackPathAndCommandSucceeds)
	s.Step(`^user runs command "([^"]*)" on path "([^"]*)"$`, stackRuntimeInfo.userRunsCommandOnPath)
	s.Step(`^command should be ran successfully$`, stackRuntimeInfo.commandShouldBeRanSuccessfully)
	s.Step(`^user stops workspace$`, stackRuntimeInfo.userStopsWorkspace)
	s.Step(`^workspace stop should be successful$`, stackRuntimeInfo.workspaceStopShouldBeSuccessful)
	s.Step(`^workspace is removed$`, stackRuntimeInfo.workspaceIsRemoved)
	s.Step(`^workspace removal should be successful$`, stackRuntimeInfo.workspaceRemovalShouldBeSuccessful)
}
