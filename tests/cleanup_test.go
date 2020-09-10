/*
 * Copyright 2020 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"process-history-cleanup/pkg"
	"process-history-cleanup/pkg/camunda"
	"process-history-cleanup/pkg/configuration"
	"process-history-cleanup/tests/docker"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCleanup(t *testing.T) {
	t.Run("with batch", testCleanup(2, 3, true, false))
}

func TestCleanupLocalFilter(t *testing.T) {
	t.Run("with batch", testCleanup(2, 3, true, false))
}

func TestCleanupLong(t *testing.T) {
	t.Skip()
	t.Run("with batch", testCleanup(500, 10000, false, true))
}

func TestCleanupLongLocalFilter(t *testing.T) {
	t.Skip()
	t.Run("with batch", testCleanup(500, 10000, false, true))
}

func TestCleanupExtraLong(t *testing.T) {
	t.Skip()
	t.Run("with batch", testCleanup(500, 100000, false, true))
}

func TestCleanupExtraLongLocalFilter(t *testing.T) {
	t.Skip()
	t.Run("with batch", testCleanup(500, 100000, false, true))
}

func testCleanup(batchSize int, deleteCount int, expectSurvivor bool, filterLocally bool) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		defer wg.Wait()
		defer cancel()

		pgip, err := docker.Postgres(ctx, wg)
		if err != nil {
			t.Error(err)
			return
		}

		camundaUrl, err := docker.Camunda(ctx, wg, pgip)

		processId := ""
		t.Run("create process", testCreateProcess(camundaUrl, &processId))

		t.Run("start process n times", testStartProcesses(camundaUrl, processId, deleteCount))
		t.Run("check n created instances", testRunCheck(camundaUrl, deleteCount))

		t.Run("run cleanup 10m", testRunCleanup(camundaUrl, "10m", batchSize, filterLocally))
		t.Run("check after 10m cleanup", testRunCheck(camundaUrl, deleteCount))

		time.Sleep(2 * time.Second)
		t.Run("start process 1 times", testStartProcesses(camundaUrl, processId, 1))
		t.Run("check 1 created instance", testRunCheck(camundaUrl, deleteCount+1))
		time.Sleep(1 * time.Second)

		expectedCount := 0
		if expectSurvivor {
			expectedCount = 1
		}
		t.Run("run cleanup 2s", testRunCleanup(camundaUrl, "2s", batchSize, filterLocally))
		t.Run("check 2s cleanup", testRunCheck(camundaUrl, expectedCount))
	}
}

func testRunCleanup(camundaUrl string, maxAge string, batchSize int, filterLocally bool) func(t *testing.T) {
	return func(t *testing.T) {
		err := pkg.RunCleanup(&configuration.ConfigStruct{
			EngineUrl:     camundaUrl,
			MaxAge:        maxAge,
			BatchSize:     batchSize,
			FilterLocally: filterLocally,
			Location:      "Europe/Berlin",
		})
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testRunCheck(camundaUrl string, expectedCount int) func(t *testing.T) {
	return func(t *testing.T) {
		count, err := camunda.New(&configuration.ConfigStruct{
			EngineUrl: camundaUrl,
		}).ListHistoryCount(true)
		if err != nil {
			t.Error(err)
			return
		}
		if count.Count != int64(expectedCount) {
			t.Error(expectedCount, count.Count)
			return
		}
	}
}

func testStartProcesses(url string, id string, count int) func(t *testing.T) {
	return func(t *testing.T) {
		for i := 0; i < count; i++ {
			testStartProcess(url, id)(t)
		}
	}
}

func testStartProcess(url string, id string) func(t *testing.T) {
	return func(t *testing.T) {
		err := startProcess(url, id)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func testCreateProcess(url string, id *string) func(t *testing.T) {
	return func(t *testing.T) {
		var err error
		*id, err = deployProcess(url, "test", createBlankProcess(), "<svg/>", "owner", "test")
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func createBlankProcess() string {
	templ := `<bpmn:definitions xmlns:xsi='http://www.w3.org/2001/XMLSchema-instance' xmlns:bpmn='http://www.omg.org/spec/BPMN/20100524/MODEL' xmlns:bpmndi='http://www.omg.org/spec/BPMN/20100524/DI' xmlns:dc='http://www.omg.org/spec/DD/20100524/DC' id='Definitions_1' targetNamespace='http://bpmn.io/schema/bpmn'><bpmn:process id='PROCESSID' isExecutable='true'><bpmn:startEvent id='StartEvent_1'/></bpmn:process><bpmndi:BPMNDiagram id='BPMNDiagram_1'><bpmndi:BPMNPlane id='BPMNPlane_1' bpmnElement='PROCESSID'><bpmndi:BPMNShape id='_BPMNShape_StartEvent_2' bpmnElement='StartEvent_1'><dc:Bounds x='173' y='102' width='36' height='36'/></bpmndi:BPMNShape></bpmndi:BPMNPlane></bpmndi:BPMNDiagram></bpmn:definitions>`
	return strings.Replace(templ, "PROCESSID", "id_"+strconv.FormatInt(time.Now().Unix(), 10), 1)
}

func deployProcess(engineUrl string, name string, xml string, svg string, owner string, source string) (definitionId string, err error) {
	responseWrapper, err := deployProcessXml(engineUrl, name, xml, svg, owner, source)
	if err != nil {
		log.Println("ERROR: unable to decode process engine deployment response", err)
		return definitionId, err
	}
	definitions, ok := responseWrapper["deployedProcessDefinitions"].(map[string]interface{})
	if !ok {
		log.Println("ERROR: unable to interpret process engine deployment response", responseWrapper)
		return definitionId, errors.New("unable to interpret process engine deployment response")
	}
	for id, _ := range definitions {
		definitionId = id
	}
	if !ok {
		log.Println("ERROR: unable to interpret process engine deployment response", responseWrapper)
		return definitionId, errors.New("unable to interpret process engine deployment response")
	}
	if definitionId == "" {
		err = errors.New("process-engine didnt deploy process: " + xml)
	}
	return
}

func deployProcessXml(engineUrl string, name string, xml string, svg string, owner string, source string) (result map[string]interface{}, err error) {
	result = map[string]interface{}{}
	boundary := "---------------------------" + time.Now().String()
	b := strings.NewReader(buildPayLoad(name, xml, svg, boundary, owner, source))
	resp, err := http.Post(engineUrl+"/engine-rest/deployment/create", "multipart/form-data; boundary="+boundary, b)
	if err != nil {
		log.Println("ERROR: request to processengine ", err)
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

func buildPayLoad(name string, xml string, svg string, boundary string, owner string, deploymentSource string) string {
	segments := []string{}
	if deploymentSource == "" {
		deploymentSource = "sepl"
	}

	segments = append(segments, "Content-Disposition: form-data; name=\"data\"; "+"filename=\""+name+".bpmn\"\r\nContent-Type: text/xml\r\n\r\n"+xml+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"diagram\"; "+"filename=\""+name+".svg\"\r\nContent-Type: image/svg+xml\r\n\r\n"+svg+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-name\"\r\n\r\n"+name+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"deployment-source\"\r\n\r\n"+deploymentSource+"\r\n")
	segments = append(segments, "Content-Disposition: form-data; name=\"tenant-id\"\r\n\r\n"+owner+"\r\n")

	return "--" + boundary + "\r\n" + strings.Join(segments, "--"+boundary+"\r\n") + "--" + boundary + "--\r\n"
}

func startProcess(engineUrl string, processDefinitionId string) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(map[string]string{})
	if err != nil {
		return
	}

	resp, err := http.Post(engineUrl+"/engine-rest/process-definition/"+url.QueryEscape(processDefinitionId)+"/start", "application/json", b)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("error on process start (status != 200)")
		return
	}
	return
}
