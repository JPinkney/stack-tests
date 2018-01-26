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

package util

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
	Command []Command           `json:"commands,omitempty"`
	Name    string              `json:"name"`
}

type Workspace2 struct {
	ID string `json:"id"`
}

type StackConfigInfo struct {
	WorkspaceConfig interface{}
	Project         interface{}
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
	Name        string           `json:"name"`
	Source      SampleSourceType `json:"source"`
	Commands    []Command        `json:"commands"`
	Tags        []string         `json:"tags"`
	Path        string           `json:"path"`
	ProjectType string           `json:"projectType"`
}

type WorkspaceConfig struct {
	EnvironmentConfig EnvironmentConfig   `json:"environments"`
	Name              string              `json:"name"`
	DefaultEnv        string              `json:"defaultEnv"`
	Description       interface{}         `json:"description,omitempty"`
	Commands          []Command           `json:"commands"`
	Source            WorkspaceSourceType `json:"source,omitempty"`
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

type Che5RuntimeStruct struct {
	Runtime Che5Machine `json:"runtime"`
}

type Che5Machine struct {
	Machines []Che5Runtime `json:"machines"`
}

type Che5Runtime struct {
	Runtime Servers `json:"runtime"`
}

type Servers struct {
	Servers map[string]ServerURL `json:"servers"`
}

type ServerURL struct {
	URL string `json:"url"`
	Ref string `json:"ref,omitempty"`
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

type CheAPI struct {
	CheAPIEndpoint string
	WorkspaceID    string
	ExecAgentURL   string
	WSAgentURL     string
	PID            int
	StackName      string
}

//Helper functions
func (c *CheAPI) GetJSON(url string) []byte {

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
func (c *CheAPI) GetExecLogs(Pid int, execAgentURL string) LogArray {
	jsonData := c.GetJSON(execAgentURL + "/" + strconv.Itoa(Pid) + "/logs")
	var data LogArray
	jsonErr := json.Unmarshal(jsonData, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func (c *CheAPI) GetCommandExitCode(Pid int) ProcessStruct {
	jsonData := c.GetJSON(c.ExecAgentURL + "/" + strconv.Itoa(Pid))
	var data ProcessStruct
	jsonErr := json.Unmarshal(jsonData, &data)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return data
}

func (c *CheAPI) PostCommandToWorkspace(sampleCommand Command) int {
	sampleCommandMarshalled, _ := json.MarshalIndent(sampleCommand, "", "    ")
	req, err := http.NewRequest("POST", c.ExecAgentURL, bytes.NewBufferString(string(sampleCommandMarshalled)))
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
func (c *CheAPI) AddSamplesToProject(sample []Sample) error {

	var sampleArray []interface{}
	for _, workspaceSample := range sample {
		sampleArray = append(sampleArray, workspaceSample)
	}

	marshalled, _ := json.MarshalIndent(sampleArray, "", "    ")
	req, err := http.NewRequest("POST", c.WSAgentURL+"/project/batch", bytes.NewBufferString(string(marshalled)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Could not complete the http request: %v", err)
	}

	defer resp.Body.Close()

	return nil
}

func (c *CheAPI) GetNumberOfProjects() (int, error) {

	projectData := c.GetJSON(c.WSAgentURL + "/project")

	var data []Sample
	jsonErr := json.Unmarshal(projectData, &data)
	if jsonErr != nil {
		return -1, fmt.Errorf("Could not unmarshall data into []Sample: %v", jsonErr)
	}

	return len(data), nil
}

func (c *CheAPI) BlockWorkspaceUntilStarted(workspaceID string) error {
	workspaceStatus, err := c.GetWorkspaceStatusByID(workspaceID)
	if err != nil {
		return err
	}
	for workspaceStatus.WorkspaceStatus == "STARTING" {
		time.Sleep(30 * time.Second)
		workspaceStatus, err = c.GetWorkspaceStatusByID(workspaceID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CheAPI) BlockWorkspaceUntilStopped(workspaceID string) error {
	workspaceStatus, err := c.GetWorkspaceStatusByID(workspaceID)
	if err != nil {
		return err
	}
	//Workspace hasn't quite shut down due to speed
	for workspaceStatus.WorkspaceStatus == "SNAPSHOTTING" || workspaceStatus.WorkspaceStatus == "STOPPING" {
		time.Sleep(15 * time.Second)
		workspaceStatus, err = c.GetWorkspaceStatusByID(workspaceID)
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

func (c *CheAPI) GetHTTPAgents(workspaceID string) (Agent, error) {
	var agents Agent

	//Now we need to get the workspace installers and then unmarshall
	runtimeData := c.GetJSON(c.CheAPIEndpoint + "/workspace/" + workspaceID)

	var data RuntimeStruct
	jsonErr := json.Unmarshal(runtimeData, &data)
	if jsonErr != nil {
		//fmt.Printf("Could not unmarshall data into RuntimeStruct: %v", jsonErr)
	}

	var data2 Che5RuntimeStruct
	jsonErr2 := json.Unmarshal(runtimeData, &data2)
	if jsonErr2 != nil {
		//fmt.Printf("Could not unmarshall data into Che5Runtime: %v", jsonErr2)
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

	for index := range data2.Runtime.Machines {
		for _, server := range data2.Runtime.Machines[index].Runtime.Servers {
			//fmt.Printf("SERVER IS: %s", server)
			if server.Ref == "exec-agent" {
				agents.execAgentURL = server.URL + "/process"
			}

			if server.Ref == "wsagent" {
				agents.wsAgentURL = server.URL
			}
		}
	}

	return agents, nil
}

func (c *CheAPI) StartWorkspace(workspaceConfiguration interface{}, stackID string) (Workspace2, error) {

	a := Post{Environments: workspaceConfiguration, Namespace: "che", Name: stackID + "-stack-test", DefaultEnv: "default"}
	marshalled, _ := json.MarshalIndent(a, "", "    ")
	re := regexp.MustCompile(",[\\n|\\s]*\"com.redhat.bayesian.lsp\"")
	noBayesian := re.ReplaceAllString(string(marshalled), "")

	//fmt.Printf("%s\n", noBayesian)

	req, err := http.NewRequest("POST", c.CheAPIEndpoint+"/workspace?start-after-create=true", bytes.NewBufferString(noBayesian))

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

func (c *CheAPI) GetWorkspaceStatusByID(workspaceID string) (WorkspaceStatus, error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	var data WorkspaceStatus

	buf2 := new(bytes.Buffer)
	url := c.CheAPIEndpoint + "/workspace/" + workspaceID
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

	//fmt.Printf("Trying to understand wtf is happening")
	//fmt.Printf("%v", string(body))

	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		return data, fmt.Errorf("Could not unmarshal contents into WorkspaceStatus: %v", jsonErr)
	}

	return data, nil
}

func (c *CheAPI) CheckWorkspaceDeletion(workspaceID string) (int, error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	buf2 := new(bytes.Buffer)
	url := c.CheAPIEndpoint + "/workspace/" + workspaceID
	req, err := http.NewRequest(http.MethodGet, url, buf2)
	res, _ := client.Do(req)
	if err != nil {
		return -1, err
	}
	return res.StatusCode, nil
}

func (c *CheAPI) StopWorkspace(workspaceID string) error {
	url := c.CheAPIEndpoint + "/workspace/" + workspaceID + "/runtime"
	req, err := http.NewRequest("DELETE", url, bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	resp.Body.Close()

	c.BlockWorkspaceUntilStopped(workspaceID)

	return nil
}

func (c *CheAPI) RemoveWorkspace(workspaceID string) error {
	url := c.CheAPIEndpoint + "/workspace/" + workspaceID
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

func (c *CheAPI) SetAgentsURL(agents Agent) {
	c.WSAgentURL = agents.wsAgentURL
	c.ExecAgentURL = agents.execAgentURL
}

func (c *CheAPI) SetWorkspaceID(workspaceID string) {
	c.WorkspaceID = workspaceID
}

func (c *CheAPI) SetStackName(stackName string) {
	c.StackName = stackName
}
