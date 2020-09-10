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

package camunda

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"
)

func (this *Camunda) ListHistory(limit string, offset string, sortby string, sortdirection string, finished bool) (result HistoricProcessInstances, err error) {
	params := url.Values{
		"maxResults":  []string{limit},
		"firstResult": []string{offset},
		"sortBy":      []string{sortby},
		"sortOrder":   []string{sortdirection},
	}
	if finished {
		params["finished"] = []string{"true"}
	} else {
		params["unfinished"] = []string{"true"}
	}

	path := "/engine-rest/history/process-instance?" + params.Encode()
	req, err := http.NewRequest("GET", this.config.EngineUrl+path, nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("%w %v %v", ErrUnexpectedResponse, resp.Status, string(buf))
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		err = fmt.Errorf("%w %v", ErrUnexpectedResponse, err.Error())
		return result, err
	}
	return result, err
}

func (this *Camunda) ListHistoryFinishedBefore(limit string, offset string, sortby string, sortdirection string, finished bool, before time.Time) (result HistoricProcessInstances, err error) {
	params := url.Values{
		"maxResults":     []string{limit},
		"firstResult":    []string{offset},
		"sortBy":         []string{sortby},
		"sortOrder":      []string{sortdirection},
		"finishedBefore": []string{before.Format(CamundaTimeFormat)},
	}
	if finished {
		params["finished"] = []string{"true"}
	} else {
		params["unfinished"] = []string{"true"}
	}

	path := "/engine-rest/history/process-instance?" + params.Encode()
	req, err := http.NewRequest("GET", this.config.EngineUrl+path, nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("%w %v %v", ErrUnexpectedResponse, resp.Status, string(buf))
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		err = fmt.Errorf("%w %v", ErrUnexpectedResponse, err.Error())
		return result, err
	}
	return result, err
}

func (this *Camunda) ListHistoryCount(finished bool) (result Count, err error) {
	params := url.Values{}
	if finished {
		params["finished"] = []string{"true"}
	} else {
		params["unfinished"] = []string{"true"}
	}

	path := "/engine-rest/history/process-instance/count?" + params.Encode()
	req, err := http.NewRequest("GET", this.config.EngineUrl+path, nil)
	if err != nil {
		return result, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("%w %v %v", ErrUnexpectedResponse, resp.Status, string(buf))
		return result, err
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		err = fmt.Errorf("%w %v", ErrUnexpectedResponse, err.Error())
		return result, err
	}
	return result, err
}
