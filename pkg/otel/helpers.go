// Copyright © 2021 The Knative Authors
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

package otel

import "strings"

type FullName struct {
	Type string
	Name string // might be empty
}

func ParseFullName(n string) FullName {
	parts := strings.Split(n, "/")
	fn := FullName{Type: parts[0]}
	if len(parts) > 1 {
		fn.Name = parts[1]
	}
	return fn

}

// HasType returns true when there is at least one entry with a full name prefixed with the given type
func HasType(entries map[string]interface{}, typ string) bool {
	for fullname := range entries {
		fn := ParseFullName(fullname)

		if fn.Type == typ {
			return true
		}
	}
	return false
}

func FindExporterInServiceByType(config CollectorConfig, typ string) string {
	exporters := config.Service.Pipelines.Traces.Exporters
	for _, exporter := range exporters {
		fn := ParseFullName(exporter)
		if fn.Type == typ {
			return exporter
		}
	}
	return ""

}
