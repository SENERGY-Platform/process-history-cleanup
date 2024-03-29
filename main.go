/*
 * Copyright 2019 InfAI (CC SES)
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

package main

import (
	"flag"
	"github.com/SENERGY-Platform/process-history-cleanup/pkg"
	"github.com/SENERGY-Platform/process-history-cleanup/pkg/configuration"
	"log"
	"time"
)

func main() {
	time.Sleep(5 * time.Second) //wait for routing tables in cluster

	confLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	config, err := configuration.Load(*confLocation)
	if err != nil {
		log.Fatal("ERROR: unable to load config ", err)
	}

	err = pkg.RunCleanup(config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Interval != "" && config.Interval != "-" {
		interval, err := time.ParseDuration(config.Interval)
		if err != nil {
			log.Fatal(err)
		}
		ticker := time.NewTicker(interval)
		for range ticker.C {
			err = pkg.RunCleanup(config)
			if err != nil {
				log.Println(err)
			}
		}
	}

}
