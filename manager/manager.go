// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"time"

	"github.com/AliyunContainerService/kube-eventer/core"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Last time of eventer housekeep since unix epoch in seconds
	lastHousekeepTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "eventer",
			Subsystem: "manager",
			Name:      "last_time_seconds",
			Help:      "Last time of eventer housekeep since unix epoch in seconds.",
		})

	// Time of latest scrape operation
	LatestScrapeTime = time.Now()
)

func init() {
	prometheus.MustRegister(lastHousekeepTimestamp)
}

type Manager interface {
	Start()
	Stop()
}

type realManager struct {
	source    core.EventSource
	sink      core.EventSink
	frequency time.Duration
	stopChan  chan struct{}
}

func NewManager(source core.EventSource, sink core.EventSink, frequency time.Duration) (Manager, error) {
	manager := realManager{
		source:    source,
		sink:      sink,
		frequency: frequency,
		stopChan:  make(chan struct{}),
	}

	return &manager, nil
}

func (rm *realManager) Start() {
	go rm.Housekeep()
}

func (rm *realManager) Stop() {
	rm.stopChan <- struct{}{}
}

func (rm *realManager) Housekeep() {
	for {
		// Try to invoke housekeep at fixed time.
		now := time.Now()
		start := now.Truncate(rm.frequency)
		end := start.Add(rm.frequency)
		timeToNextSync := end.Sub(now)

		select {
		case <-time.After(timeToNextSync):
			rm.housekeep()
		case <-rm.stopChan:
			rm.sink.Stop()
			return
		}
	}
}

func (rm *realManager) housekeep() {
	defer func() {
		lastHousekeepTimestamp.Set(float64(time.Now().Unix()))
	}()

	LatestScrapeTime = time.Now()

	// No parallelism. Assumes that the events are pushed to Heapster. Add parallelism
	// when this stops to be true.
	events := rm.source.GetNewEvents()
	klog.V(0).Infof("Exporting %d events", len(events.Events))
	alterEvents := &core.EventBatch{Timestamp: events.Timestamp, Events: nil}
	if len(events.Events) > 0 {
		for _, event := range events.Events {
			// 将结构体转换为 JSON
			jsonData, err := json.Marshal(event)
			if err != nil {
				klog.V(0).Infof("events json Marshal Error:", err)
			}

			// 将 JSON 转换为字符串并打印
			fmt.Println(string(jsonData))
			if event.Type == "Warning" {
				alterEvents.Events = append(alterEvents.Events, event)
			}
		}
		rm.sink.ExportEvents(alterEvents)
	}
}
