package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

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

type LogArray []struct {
	Kind int       `json:"kind"`
	Time time.Time `json:"time"`
	Text string    `json:"text"`
}

type WorkspaceTableItem struct {
	Stack       string
	ProjectName string
	Cmd         string
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

type cheRunner struct {
	cheAPIEndpoint string
}

//Helper functions
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

//ExecAgent Processes
func getExecLogs(Pid int, execAgentURL string) LogArray {
	jsonData := getJSON(execAgentURL + "/" + strconv.Itoa(Pid) + "/logs")
	var data LogArray
	jsonErr := json.Unmarshal(jsonData, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func getCommandExitCode(Pid int, execAgentURL string) ProcessStruct {
	jsonData := getJSON(execAgentURL + "/" + strconv.Itoa(Pid))
	var data ProcessStruct
	jsonErr := json.Unmarshal(jsonData, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func postCommandToWorkspace(workspaceID, execAgentURL string, sampleCommand string, samplePath string) int {
	req, err := http.NewRequest("POST", execAgentURL, bytes.NewBufferString(sampleCommand))
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

	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		panic(err.Error())
	}

	defer resp.Body.Close()

	return data.Pid

}

//WSAgent Processes
func addSampleToProject(wsAgentURL string, sample []WorkspaceSample) error {

	var sampleArray []interface{}
	for _, workspaceSample := range sample {
		sampleArray = append(sampleArray, workspaceSample.Sample)
	}

	marshalled, _ := json.MarshalIndent(sampleArray, "", "    ")
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

func (c *cheRunner) blockWorkspaceUntilStarted(workspaceID string) error {
	workspaceStatus, err := c.getWorkspaceStatusByID(workspaceID)
	if err != nil {
		return err
	}
	for workspaceStatus.WorkspaceStatus == "STARTING" {
		time.Sleep(30 * time.Second)
		workspaceStatus, err = c.getWorkspaceStatusByID(workspaceID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cheRunner) blockWorkspaceUntilStopped(workspaceID string) error {
	workspaceStatus, err := c.getWorkspaceStatusByID(workspaceID)
	if err != nil {
		return err
	}
	//Workspace hasn't quite shut down due to speed
	for workspaceStatus.WorkspaceStatus == "SNAPSHOTTING" {
		time.Sleep(15 * time.Second)
		workspaceStatus, err = c.getWorkspaceStatusByID(workspaceID)
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

func (c *cheRunner) getHTTPAgents(workspaceID string) (Agent, error) {
	var agents Agent

	//Now we need to get the workspace installers and then unmarshall
	runtimeData := getJSON(c.cheAPIEndpoint + "/workspace/" + workspaceID)

	//fmt.Printf(string(runtimeData))
	var data RuntimeStruct
	jsonErr := json.Unmarshal(runtimeData, &data)
	if jsonErr != nil {
		return agents, fmt.Errorf("Could not unmrshall data into RuntimeStruct: %v", jsonErr)
	}

	for key := range data.Runtime.Machines {
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

func (c *cheRunner) startWorkspace(workspaceConfiguration StackConfigInfo, sample interface{}, stackID string) (Workspace2, error) {
	workspaceConfig := workspaceConfiguration.Config.EnvironmentConfig

	a := Post{Environments: workspaceConfig, Namespace: "che", Name: stackID + "-stack-test", DefaultEnv: "default"}
	marshalled, _ := json.MarshalIndent(a, "", "    ")
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")

	req, err := http.NewRequest("POST", c.cheAPIEndpoint+"/workspace?start-after-create=true", bytes.NewBufferString(noBayesian))

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

func (c *cheRunner) getWorkspaceStatusByID(workspaceID string) (WorkspaceStatus, error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	var data WorkspaceStatus

	buf2 := new(bytes.Buffer)
	url := c.cheAPIEndpoint + "/workspace/" + workspaceID
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

func (c *cheRunner) stopWorkspace(workspaceID string) error {
	url := c.cheAPIEndpoint + "/workspace/" + workspaceID + "/runtime"
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	resp.Body.Close()

	c.blockWorkspaceUntilStopped(workspaceID)

	return nil
}

func (c *cheRunner) removeWorkspace(workspaceID string) error {
	url := c.cheAPIEndpoint + "/workspace/" + workspaceID
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer resp.Body.Close()

	return nil
}
