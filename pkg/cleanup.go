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

package pkg

import (
	"errors"
	"log"
	"process-history-cleanup/pkg/camunda"
	"process-history-cleanup/pkg/configuration"
	"strconv"
	"time"
	_ "time/tzdata"
)

func RunCleanup(config configuration.Config) (err error) {
	maxAge, err := time.ParseDuration(config.MaxAge)
	if err != nil {
		return err
	}
	if config.BatchSize > 0 {
		return runCleanup(camunda.New(config), maxAge, config.BatchSize, config.FilterLocally)
	} else {
		return errors.New("expect batch size > 0")
	}
}

func runCleanup(camunda Camunda, maxAge time.Duration, batchSize int, filterLocally bool) (err error) {
	finished := false
	for !finished {
		if filterLocally {
			finished, err = runCleanupBatch(camunda, maxAge, batchSize)
		} else {
			finished, err = runCleanupBatchV2(camunda, maxAge, batchSize)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func runCleanupBatchV2(camundaEngine Camunda, maxAge time.Duration, batchSize int) (finished bool, err error) {
	//we sort so that the old process instances will be processed first
	//if this instance is younger than the maxAge than all following instances are younger too
	//all entries will be deleted until we find one that is younger than the max age
	//this means the offset may be 0 in each batch
	historyInstances, err := camundaEngine.ListHistoryFinishedBefore(strconv.Itoa(batchSize), "0", "endTime", "asc", true, time.Now().Add(-maxAge))
	if err != nil {
		return true, err
	}

	for _, instance := range historyInstances {
		log.Println("delete " + instance.Id)
		err = camundaEngine.RemoveProcessInstanceHistory(instance.Id)
		if err != nil {
			return true, err
		}
	}
	return len(historyInstances) != batchSize, nil
}

func runCleanupBatch(camundaEngine Camunda, maxAge time.Duration, batchSize int) (finished bool, err error) {
	//we sort so that the old process instances will be processed first
	//if this instance is younger than the maxAge than all following instances are younger too
	//all entries will be deleted until we find one that is younger than the max age
	//this means the offset may be 0 in each batch
	historyInstances, err := camundaEngine.ListHistory(strconv.Itoa(batchSize), "0", "endTime", "asc", true)
	if err != nil {
		return true, err
	}

	for _, instance := range historyInstances {
		endTime, err := time.Parse(camunda.CamundaTimeFormat, instance.EndTime)
		if err != nil {
			log.Println("WARNING: unable to parse end time", instance.EndTime, err)
			err = nil
			continue
		}
		if time.Since(endTime) > maxAge {
			log.Println("delete " + instance.Id)
			err = camundaEngine.RemoveProcessInstanceHistory(instance.Id)
			if err != nil {
				return true, err
			}
		} else {
			return true, nil
		}
	}
	return len(historyInstances) != batchSize, nil
}
