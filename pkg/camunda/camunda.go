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
	"github.com/SENERGY-Platform/process-history-cleanup/pkg/configuration"
	"log"
	"time"
)

type Camunda struct {
	config   configuration.Config
	location *time.Location
}

func New(config configuration.Config) *Camunda {
	location, err := time.LoadLocation(config.Location)
	if err != nil {
		log.Println("unable to load location")
		location, _ = time.LoadLocation("Europe/Berlin")
	}
	return &Camunda{config: config, location: location}
}
