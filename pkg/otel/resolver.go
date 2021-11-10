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

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Resolve resolves the given endpoint to a real zipkin endpoint
func ResolveZipkin(ctx context.Context, endpoint string, restcfg *rest.Config) (string, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	parts := regexp.MustCompile("[.:]").Split(url.Host, -1)
	if len(parts) < 2 {
		return "", fmt.Errorf("malformed endpoint %q", endpoint)
	}

	// let's assume otel is installed in the cluster
	svcName := parts[0]
	svcNamespace := parts[1]

	kubeclient, err := kubernetes.NewForConfig(restcfg)
	if err != nil {
		return "", err
	}

	cm, err := kubeclient.CoreV1().ConfigMaps(svcNamespace).Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		// wrong assumption. bail out
		return "", err
	}

	collectorYAML, ok := cm.Data["collector.yaml"]
	if !ok {
		return "", errors.New("missing collector.yaml key")
	}

	var collectorCfg CollectorConfig
	err = yaml.Unmarshal([]byte(collectorYAML), &collectorCfg)
	if err != nil {
		return "", err
	}

	// Look for zipkin receiver (since Knative Eventing only support sending Zipkin traces)
	if !HasType(collectorCfg.Receivers, "zipkin") {
		return "", errors.New("OpenTelemetry collector not receiving Zipkin traces")
	}

	// Check traces are exported to a zipkin instance
	zipkinName := FindExporterInServiceByType(collectorCfg, "zipkin")
	if zipkinName == "" {
		return "", errors.New("OpenTelemetry collector does not export traces to Zipkin")
	}

	zipkinConfig, ok := collectorCfg.Exporters[zipkinName]
	if !ok {
		return "", errors.New("OpenTelemetry collector does not export traces to Zipkin (invalid configuration)")
	}

	resolved, ok := zipkinConfig["endpoint"]
	if !ok {
		return "", errors.New("OpenTelemetry collector does not export traces to Zipkin (missing Zipkin endpoint)")
	}

	return resolved.(string), nil
}
