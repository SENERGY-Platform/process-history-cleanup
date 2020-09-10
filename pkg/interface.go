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
	"process-history-cleanup/pkg/camunda"
	"time"
)

type Camunda interface {
	ListHistoryFinishedBefore(limit string, offset string, sortby string, sortdirection string, finished bool, before time.Time) (result camunda.HistoricProcessInstances, err error)
	ListHistory(limit string, offset string, sortby string, sortdirection string, finished bool) (result camunda.HistoricProcessInstances, err error)
	RemoveProcessInstanceHistory(id string) (err error)
}