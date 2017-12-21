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
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

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

func getExecAgentHTTP(workspaceID string) Agent {
	//Now we need to get the workspace installers and then unmarshall
	runtimeData := getJSON(fullyQualifiedEndpoint + "/workspace/" + workspaceID)
	//fmt.Printf(string(runtimeData))
	var data RuntimeStruct
	jsonErr := json.Unmarshal(runtimeData, &data)
	if jsonErr != nil {
		fmt.Printf("%v", jsonErr)
	}

	//fmt.Printf("%v", data)

	var agents Agent
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

	return agents
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

func continuouslyCheckCommandExitCode(Pid int, execAgentURL string) {
	runCommand := checkCommandExitCode(Pid, execAgentURL)
	time.Sleep(15 * time.Second)
	checkExecStatus(Pid, runCommand.ExitCode, execAgentURL)
	for runCommand.Alive == true {
		time.Sleep(15 * time.Second)
		runCommand = checkCommandExitCode(Pid, execAgentURL)
		checkExecStatus(Pid, runCommand.ExitCode, execAgentURL)
	}

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

type Stack struct {
	Name             string
	ImageName        string
	Sample           string
	Cmd              string
	ExpectedOutput   string
	Output           string
	SampleFolderName string
}

func generateExampleTables(stackData []Workspace, samples []Sample, tag string) []WorkspaceSample {
	var tableElements []WorkspaceSample

	for _, stackElement := range stackData {
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
					tableElements = append(tableElements, WorkspaceSample{
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
				tableElements = append(tableElements, WorkspaceSample{
					Command:    commandList,
					Config:     stackElement.Config,
					ID:         stackElement.ID,
					Sample:     sampleElement,
					SamplePath: sampleElement.Path,
				})
			}
		}
	}
	return tableElements
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

func (stack *Stack) weCheckRunCommandsAsDefaultUser() error {

	stack.stopDockerContainer() //Stop the container just in case its running

	runCommand := "docker run -i --rm --name " + stack.Name + " -v /tmp/" + stack.Name + ":/projects/" + stack.Name + "/ " + stack.ImageName + " sh -c"
	runCommandSplitWithoutNode := strings.Split(runCommand, " ")

	cmdReplacer := strings.Replace(stack.Cmd, "${current.project.path}", stack.Sample, -1)
	cmdReplacer = strings.Replace(cmdReplacer, "${GAE}", "/home/user/google_appengine", -1)
	cmdReplacer = strings.Replace(cmdReplacer, "$TOMCAT_HOME", "/home/user/tomcat8", -1)

	runCommandSplitWithNode := append(runCommandSplitWithoutNode, strings.Split(cmdReplacer, " ")...)
	stdout, err := execWithPiping(runCommandSplitWithNode)
	if err != nil {
		return fmt.Errorf("Docker run has failed: %s", err)
	}

	stack.ExpectedOutput = stdout

	stack.stopDockerContainer() //Stop the container just in case its running

	return nil
}

func (stack *Stack) weCheckRunCommandsAsArbitraryUser() error {

	stack.stopDockerContainer() //Stop the container just in case its running

	runCommand := "docker run -i --rm --name " + stack.Name + " --user 15151515 -v /tmp/" + stack.Name + ":/projects/" + stack.Name + "/ " + stack.ImageName + " sh -c"
	runCommandSplitWithoutNode := strings.Split(runCommand, " ")

	cmdReplacer := strings.Replace(stack.Cmd, "${current.project.path}", stack.Sample, -1)
	cmdReplacer = strings.Replace(cmdReplacer, "${GAE}", "/home/user/google_appengine", -1)
	cmdReplacer = strings.Replace(cmdReplacer, "$TOMCAT_HOME", "/home/user/tomcat8", -1)

	runCommandSplitWithNode := append(runCommandSplitWithoutNode, strings.Split(cmdReplacer, " ")...)

	stdout, err := execWithPiping(runCommandSplitWithNode)
	if err != nil {
		return fmt.Errorf("Docker run has failed: %s", err)
	}

	stack.ExpectedOutput = stdout

	stack.stopDockerContainer() //Stop the container just in case its running

	return nil
}

func (stack *Stack) weCheckRunMainBinaryFromBashAsDefaultUser() error {
	stack.stopDockerContainer() //Stop the container just in case its running

	runCommand := "docker run -i --rm --name " + stack.Name + " " + stack.ImageName + " sh -c"
	runCommandSplitWithoutNode := strings.Split(runCommand, " ")

	cmdReplacer := strings.Replace(stack.Cmd, "${current.project.path}", stack.SampleFolderName, -1)
	cmdReplacer = strings.Replace(cmdReplacer, "${GAE}", "/home/user/google_appengine", -1)
	cmdReplacer = strings.Replace(cmdReplacer, "$TOMCAT_HOME", "/home/user/tomcat8", -1)

	runCommandSplitWithNode := append(runCommandSplitWithoutNode, strings.Split(cmdReplacer, " ")...)

	stdout, err := execWithPiping(runCommandSplitWithNode)
	if err != nil {
		return fmt.Errorf("Docker run has failed: %s", err)
	}

	stack.ExpectedOutput = stdout

	stack.stopDockerContainer() //Stop the container just in case its running

	return nil
}

func (stack *Stack) weCheckRunMainBinaryFromBashAsArbitraryUser() error {
	stack.stopDockerContainer() //Stop the container just in case its running

	runCommand := "docker run -i --rm --name " + stack.Name + " --user 15151515 " + stack.ImageName + " sh -c"
	runCommandSplitWithoutNode := strings.Split(runCommand, " ")
	cmdReplacer := strings.Replace(stack.Cmd, "${current.project.path}", stack.SampleFolderName, -1)
	cmdReplacer = strings.Replace(cmdReplacer, "${GAE}", "/home/user/google_appengine", -1)
	cmdReplacer = strings.Replace(cmdReplacer, "$TOMCAT_HOME", "/home/user/tomcat8", -1)

	runCommandSplitWithNode := append(runCommandSplitWithoutNode, strings.Split(cmdReplacer, " ")...)
	stdout, err := execWithPiping(runCommandSplitWithNode)
	if err != nil {
		return fmt.Errorf("Docker run has failed: %s", err)
	}

	stack.ExpectedOutput = stdout

	stack.stopDockerContainer() //Stop the container just in case its running

	return nil
}

func (stack *Stack) removeTempStackLocation() {
	runCommand := "rm -rf /tmp/" + stack.Name
	runCommandSplit := strings.Split(runCommand, " ")
	exec.Command(runCommandSplit[0], runCommandSplit[1:]...).Run()
}

func (stack *Stack) stopDockerContainer() error {
	_, dockerRunErr := exec.Command("docker", "stop", stack.Name).Output()
	if dockerRunErr != nil {
		return fmt.Errorf("Docker run has failed: %s", dockerRunErr)
	}

	_, dockerRmErr := exec.Command("docker", "rm", stack.Name).Output()
	if dockerRmErr != nil {
		return fmt.Errorf("Docker run has failed: %s", dockerRunErr)
	}
	return nil
}

func (stack *Stack) weHaveStackNameImageNameCmdExpectedOutputSampleAndSampleFolderName(name, imageName, cmd, expectedOutput, sample, sampleFolderName string) error {
	if name == "" || imageName == "" || cmd == "" {
		return fmt.Errorf("One of the args has not been set")
	}

	stack.Name = name
	stack.ImageName = imageName
	stack.Cmd = cmd
	stack.ExpectedOutput = expectedOutput
	stack.Sample = sample
	stack.SampleFolderName = sampleFolderName

	return nil
}

func (stack *Stack) weCheckExecOfMainBinaryAsDefaultUser() error {
	stack.stopDockerContainer() //Stop the container just in case its running

	_, dockerRunErr := exec.Command("docker", "run", "-d", "--name", stack.Name, stack.ImageName, "tail", "-f", "/dev/null").Output()
	if dockerRunErr != nil {
		return fmt.Errorf("Docker run has failed: %s", dockerRunErr)
	}

	gitCloneRunCommand := "docker exec -i " + stack.Name + " git clone " + stack.Sample + " " + stack.SampleFolderName
	gitCloneRunCommandSplitArgs := strings.Split(gitCloneRunCommand, " ")

	execWithPiping(gitCloneRunCommandSplitArgs)

	cmdReplacer := strings.Replace(stack.Cmd, "${current.project.path}", stack.SampleFolderName, -1)
	cmdReplacer = strings.Replace(cmdReplacer, "${GAE}", "/home/user/google_appengine", -1)
	cmdReplacer = strings.Replace(cmdReplacer, "$TOMCAT_HOME", "/home/user/tomcat8", -1)

	runCommand := "docker exec -i " + stack.Name + " " + cmdReplacer
	runCommandSplitArgs := strings.Split(runCommand, " ")

	stdout, err := execWithPiping(runCommandSplitArgs)

	if err != nil {
		return fmt.Errorf("Docker exec failed: %s", err)
	}

	stack.Output = stdout

	dockerStopErr := stack.stopDockerContainer()
	if dockerStopErr != nil {
		return fmt.Errorf("Docker stop failed: %s", dockerStopErr)
	}

	return nil
}

func (stack *Stack) stdoutShouldBe(stdout string) error {
	if strings.Contains(stdout, stack.Output) && stack.ExpectedOutput != "" {
		return fmt.Errorf("Result was not expected. Got %s, Expected %s", stdout, stack.ExpectedOutput)
	}
	return nil
}

func (stack *Stack) weCheckExecOfMainBinaryAsArbitraryUser() error {
	stack.stopDockerContainer() //Stop the container just in case its running

	_, dockerRunErr := exec.Command("docker", "run", "-d", "--name", stack.Name, "--user", "15151515", stack.ImageName, "tail", "-f", "/dev/null").Output()
	if dockerRunErr != nil {
		return fmt.Errorf("Docker run has failed: %s", dockerRunErr)
	}

	cmdReplacer := strings.Replace(stack.Cmd, "${current.project.path}", stack.SampleFolderName, -1)
	cmdReplacer = strings.Replace(cmdReplacer, "${GAE}", "/home/user/google_appengine", -1)
	cmdReplacer = strings.Replace(cmdReplacer, "$TOMCAT_HOME", "/home/user/tomcat8", -1)
	runCommand := "docker exec -i " + stack.Name + " " + cmdReplacer
	runCommandSplitArgs := strings.Split(runCommand, " ")
	stdout, execErr := execWithPiping(runCommandSplitArgs)
	if execErr != nil {
		return fmt.Errorf("Docker exec failed: %s", execErr)
	}

	stack.ExpectedOutput = stdout

	dockerStopErr := stack.stopDockerContainer()
	if dockerStopErr != nil {
		return fmt.Errorf("Docker stop failed: %s", dockerStopErr)
	}

	return nil
}

func (stack *Stack) theImageIsBuiltAndWeHaveStackNameImageNameSampleCmdExpectedOutput(name, imageName, sample, cmd, expectedOutput string) error {
	if name == "" || imageName == "" || sample == "" || cmd == "" || expectedOutput == "" {
		return fmt.Errorf("One of the args has not been set")
	}

	stack.Name = name
	stack.ImageName = imageName
	stack.Sample = sample
	stack.Cmd = cmd
	stack.ExpectedOutput = expectedOutput

	return nil
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

//
//
//	Workspace stuff
//
//

// triggerStackStart takes in a stack configuration and starts an Eclipse Che
// workspace using that configuration
//
// Returns the http request response when starting the workspace
func triggerStackStart(workspaceConfiguration WorkspaceSample, sample interface{}) Workspace2 {
	workspaceName := workspaceConfiguration.ID
	workspaceConfig := workspaceConfiguration.Config
	test, err1 := json.Marshal(workspaceConfig)
	if err1 != nil {
		log.Fatal(err1)
	}

	jsonBytes := []byte(string(test))
	WorkspaceConfigInterface := &WorkspaceConfig{}
	json.Unmarshal(jsonBytes, WorkspaceConfigInterface)

	a := Post{Environments: WorkspaceConfigInterface.EnvironmentConfig, Namespace: "che", Name: workspaceName + "-stack-test", DefaultEnv: "default"}
	marshalled, _ := json.MarshalIndent(a, "", "    ")
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")
	req, err := http.NewRequest("POST", fullyQualifiedEndpoint+"/workspace?start-after-create=true", bytes.NewBufferString(noBayesian))
	//fmt.Print(fullyQualifiedEndpoint + "/workspace?start-after-create=true")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	//newStr := buf.String()
	//fmt.Printf(newStr)
	defer resp.Body.Close()

	//fmt.Printf(string(buf.Bytes()))

	var WorkspaceResponse Workspace2
	json.Unmarshal(buf.Bytes(), &WorkspaceResponse)

	return WorkspaceResponse
}

func tableRowGenerator(cellData Stack) *gherkin.TableRow {

	var newTableCellNode gherkin.Node
	newTableCellNode.Type = "TableCell"

	var newCell gherkin.TableCell
	newCell.Node = newTableCellNode
	newCell.Value = cellData.Name

	var newTableCellNode2 gherkin.Node
	newTableCellNode2.Type = "TableCell"

	var newCell2 gherkin.TableCell
	newCell2.Node = newTableCellNode
	newCell2.Value = cellData.ImageName

	var newTableCellNode3 gherkin.Node
	newTableCellNode3.Type = "TableCell"

	var newCell3 gherkin.TableCell
	newCell3.Node = newTableCellNode
	newCell3.Value = cellData.Cmd

	var newTableCellNode4 gherkin.Node
	newTableCellNode4.Type = "TableCell"

	var newCell4 gherkin.TableCell
	newCell4.Node = newTableCellNode
	newCell4.Value = cellData.ExpectedOutput

	var newTableCellNode5 gherkin.Node
	newTableCellNode5.Type = "TableCell"

	var newCell5 gherkin.TableCell
	newCell5.Node = newTableCellNode
	newCell5.Value = cellData.Sample

	var newCell6 gherkin.TableCell
	newCell6.Node = newTableCellNode
	newCell6.Value = cellData.SampleFolderName

	var cells []*gherkin.TableCell
	cells = append(cells, &newCell, &newCell2, &newCell3, &newCell4, &newCell5, &newCell6)

	var newRow gherkin.TableRow
	newRow.Node = gherkin.Node{Type: "TableRow"}
	newRow.Cells = cells[0:]

	return &newRow

}

func tableRowArrayGenerator(cellDataArray []Stack) []*gherkin.TableRow {

	var tableRowArray []*gherkin.TableRow

	for _, tableItem := range cellDataArray {

		newTableRow := tableRowGenerator(tableItem)
		tableRowArray = append(tableRowArray, newTableRow)

	}

	return tableRowArray
}

func (runArgsData *runArgsData) setupExamplesData(g *gherkin.Feature) {
	for _, scenario := range g.ScenarioDefinitions {
		row := scenario.(*gherkin.ScenarioOutline).Examples[0].TableBody
		newTableRow := tableRowArrayGenerator(runArgsData.Data)
		if len(newTableRow) == 1 {
			row = newTableRow
		} else {
			row = append(row, newTableRow...)
		}

		scenario.(*gherkin.ScenarioOutline).Examples[0].TableBody = row
	}
}

type runArgsData struct {
	Data []Stack
}

func FeatureContext(s *godog.Suite) {

	stackFeature := &Stack{}

	s.BeforeFeature(tableData.setupExamplesData)
	s.Step(`^we have stack name "([^"]*)" imageName "([^"]*)" cmd "([^"]*)" expectedOutput "([^"]*)" sample "([^"]*)" and sampleFolderName "([^"]*)"$`, stackFeature.weHaveStackNameImageNameCmdExpectedOutputSampleAndSampleFolderName)
	s.Step(`^we check exec of main binary as default user$`, stackFeature.weCheckExecOfMainBinaryAsDefaultUser)
	s.Step(`^stdout should be "([^"]*)"$`, stackFeature.stdoutShouldBe)
	s.Step(`^we check exec of main binary as arbitrary user$`, stackFeature.weCheckExecOfMainBinaryAsArbitraryUser)
	s.Step(`^we check run main binary from bash as default user$`, stackFeature.weCheckRunMainBinaryFromBashAsDefaultUser)
	s.Step(`^we check run main binary from bash as arbitrary user$`, stackFeature.weCheckRunMainBinaryFromBashAsArbitraryUser)
	s.Step(`^we check run commands as default user$`, stackFeature.weCheckRunCommandsAsDefaultUser)
	s.Step(`^we check run commands as arbitrary user$`, stackFeature.weCheckRunCommandsAsArbitraryUser)

}

func testSingleStack(name, imageName, cmd, expectedOutput, sample string) []Stack {
	var newSingleStackItem Stack
	newSingleStackItem.Name = name
	newSingleStackItem.ImageName = imageName
	newSingleStackItem.Cmd = cmd
	newSingleStackItem.Sample = sample
	newSingleStackItem.ExpectedOutput = expectedOutput
	goDogTableItemArray := []Stack{newSingleStackItem}
	return goDogTableItemArray
}

func testAllStacks(tag string) []WorkspaceSample {
	stackData := getJSON(eclipseStackLocation)
	var data []Workspace
	jsonErr := json.Unmarshal(stackData, &data)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	samples := getSamplesJSON(samples)
	return generateExampleTables(data, samples, tag)
}

func addSampleToProject(wsAgentURL string, sample interface{}) {
	sampleSlice := []interface{}{sample}
	marshalled, _ := json.MarshalIndent(sampleSlice, "", "    ")
	req, err := http.NewRequest("POST", wsAgentURL+"/project/batch", bytes.NewBufferString(string(marshalled)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
}

func stopWorkspace(workspaceID string) error {
	url := fullyQualifiedEndpoint + "/workspace/" + workspaceID + "/runtime"
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	resp.Body.Close()
	return nil
}

func removeWorkspace(workspaceID string) error {

	url := fullyQualifiedEndpoint + "/workspace/" + workspaceID
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	resp.Body.Close()
	return nil
}

func getWorkspaceStatusByID(workspaceID string) WorkspaceStatus {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	buf2 := new(bytes.Buffer)
	url := fullyQualifiedEndpoint + "/workspace/" + workspaceID
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

	var data WorkspaceStatus
	jsonErr := json.Unmarshal([]byte(body), &data)

	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func blockWorkspaceUntilStarted(workspaceID string) {
	workspaceStatus := getWorkspaceStatusByID(workspaceID)
	for workspaceStatus.WorkspaceStatus == "STARTING" {
		time.Sleep(30 * time.Second)
		workspaceStatus = getWorkspaceStatusByID(workspaceID)
	}
}

func blockWorkspaceUntilStopped(workspaceID string) error {
	workspaceStatus := getWorkspaceStatusByID(workspaceID)

	//Workspace hasn't quite shut down due to speed
	for workspaceStatus.WorkspaceStatus == "SNAPSHOTTING" {
		time.Sleep(15 * time.Second)
		workspaceStatus = getWorkspaceStatusByID(workspaceID)
	}

	time.Sleep(15 * time.Second)

	if workspaceStatus.WorkspaceStatus != "STOPPED" {
		return fmt.Errorf("Workspace was not stopped")
	}
	return nil
}

func TestMain(m *testing.M) {
	// singleStackTestPtr := flag.Bool("s", false, "Start Tests on a Single Stack (Optional)")

	// namePtr := flag.String("name", "", "Set a name for the Stack. Only available when -s is enabled.")
	// imageNamePtr := flag.String("image_loc", "", "Set a image name for the Stack. Only available when -s is enabled.")
	// cmdToTestPtr := flag.String("cmd", "", "Set a command to test on the Stack. Only available when -s is enabled.")
	// expectedOutputPtr := flag.String("eo", "", "Set the expected value of cmd. Only available when -s is enabled.")
	// samplePtr := flag.String("sample", "", "Set the sample project of cmd. Only available when -s is enabled.")

	// allStacksTestsPtr := flag.Bool("all", false, "Start Tests for All Stacks (Default)")
	// allStacksTestByTagPtr := flag.String("t", "?", "Start Tests for All Stacks Using a Tag (Optional)")

	// flag.Parse()

	// if *singleStackTestPtr && (*allStacksTestsPtr || *allStacksTestByTagPtr == "") {
	// 	fmt.Print("Only one of args (s, a, t) args are allowed")
	// 	os.Exit(1)
	// }

	// if *allStacksTestsPtr && *allStacksTestByTagPtr == "" {
	// 	fmt.Print("Only one of (a, t) args are allowed")
	// 	os.Exit(1)
	// }

	// if *singleStackTestPtr {
	// 	tableData.Data = testSingleStack(*namePtr, *imageNamePtr, *cmdToTestPtr, *expectedOutputPtr, *samplePtr)
	// } else if *allStacksTestsPtr || *allStacksTestByTagPtr == "" {
	// 	tableData.Data = testAllStacks(*allStacksTestByTagPtr)
	// } else {
	// 	fmt.Print("Err: Missing an argument")
	// 	os.Exit(1)
	// }

	// status := godog.RunWithOptions("godog", func(s *godog.Suite) {
	// 	FeatureContext(s)
	// }, godog.Options{
	// 	Format: "progress",
	// 	Paths:  []string{"features"},
	// })

	// start := time.Now()
	// if st := m.Run(); st > status {
	// 	status = st
	// }
	// elapsed := time.Since(start)
	// os.Exit(status)
	// fmt.Printf("go test -all took %s", elapsed)

	allStackData := testAllStacks("")

	for _, workspace := range allStackData {
		fmt.Printf("Starting workspace tests for %s\n", workspace.ID)
		workspaceStartingResp := triggerStackStart(workspace, workspace.Sample)
		blockWorkspaceUntilStarted(workspaceStartingResp.ID)
		agents := getExecAgentHTTP(workspaceStartingResp.ID)

		for agents.execAgentURL == "" || agents.wsAgentURL == "" {
			agents = getExecAgentHTTP(workspaceStartingResp.ID)
		}
		addSampleToProject(agents.wsAgentURL, workspace.Sample)

		for _, cmd := range workspace.Command {
			Pid := postCommandToWorkspace(workspaceStartingResp.ID, agents.execAgentURL, cmd, workspace.SamplePath)
			continuouslyCheckCommandExitCode(Pid, agents.execAgentURL)
		}

		stopWorkspace(workspaceStartingResp.ID)
		blockWorkspaceUntilStopped(workspaceStartingResp.ID)
		removeWorkspace(workspaceStartingResp.ID)
	}

}
