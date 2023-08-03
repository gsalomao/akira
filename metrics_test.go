// Copyright 2023 Gustavo Salomao
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

package akira

import (
	"encoding/json"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestMetricsJSON(t *testing.T) {
	golden, err := os.ReadFile("testdata/metrics.json")
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}

	var want any
	err = json.Unmarshal(golden, &want)
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}

	metrics := Metrics{}
	metrics.PacketReceived.value.Add(10)
	metrics.PacketSent.value.Add(8)
	metrics.BytesReceived.value.Add(300)
	metrics.BytesSent.value.Add(100)
	metrics.ClientsConnected.value.Add(5)
	metrics.ClientsMaximum.value.Add(8)
	metrics.PersistentSessions.value.Add(3)

	var data []byte
	data, err = json.Marshal(&metrics)
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}

	var got any
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("Unexpected error\n%v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Unexpected JSON\nwant: %v\ngot:  %v", want, got)
	}
}

func TestMetricCounterValue(t *testing.T) {
	counter := MetricCounter{}
	for i := uint64(0); i < 10; i++ {
		if counter.Value() != i {
			t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", i, counter.Value())
		}
		counter.value.Add(1)
	}
}

func TestMetricGaugeValue(t *testing.T) {
	gauge := MetricGauge{}
	for i := int64(0); i < 10; i++ {
		if gauge.Value() != i {
			t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", i, gauge.Value())
		}
		gauge.value.Add(1)
	}
	for i := int64(10); i >= 0; i-- {
		if gauge.Value() != i {
			t.Fatalf("Unexpected value\nwant: %v\ngot:  %v", i, gauge.Value())
		}
		gauge.value.Add(-1)
	}
}

func BenchmarkMetrics(b *testing.B) {
	var (
		workers = 1000
		metrics = Metrics{}
	)

	var wg sync.WaitGroup
	wg.Add(workers)
	b.ResetTimer()

	for i := 0; i < workers; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				metrics.PacketReceived.value.Add(1)
				metrics.PacketSent.value.Add(1)
				metrics.BytesReceived.value.Add(1)
				metrics.BytesSent.value.Add(1)
				metrics.ClientsConnected.value.Add(1)
				metrics.ClientsMaximum.value.Add(1)
				metrics.PersistentSessions.value.Add(1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
