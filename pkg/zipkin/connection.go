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

package zipkin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"

	"github.com/openzipkin/zipkin-go/model"
	"k8s.io/client-go/rest"
	"knative.dev/kn-plugin-trace/pkg/proxy"
)

type Connection struct {
	external bool
	proxy    proxy.Proxy

	svcName      string
	svcNamespace string
}

func Connect(endpoint string, restcfg *rest.Config) (*Connection, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	parts := regexp.MustCompile("[.:]").Split(url.Host, -1)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed endpoint %q", endpoint)
	}

	// TODO: detect whether zipkin is externally accessible, in which case there is no need to create a proxy

	proxy, err := proxy.New(restcfg)
	if err != nil {
		return nil, err
	}

	connection := Connection{
		external:     false, // TODO.
		proxy:        proxy,
		svcName:      parts[0],
		svcNamespace: parts[1],
	}

	// Check if endpoint is reachable
	_, err = connection.Services()
	if err != nil {
		return nil, err
	}

	return &connection, nil
}

func (c *Connection) Services() (ServicesResponse, error) {
	resp, err := c.proxy.Get(c.svcName, c.svcNamespace, "api/v2/services")
	if err != nil {
		return nil, err
	}

	var services ServicesResponse
	err = json.Unmarshal([]byte(resp), &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (c *Connection) Spans(serviceName string) ([][]model.SpanModel, error) {
	traces, err := c.proxy.Get(c.svcName, c.svcNamespace, fmt.Sprintf("api/v2/traces?serviceName=%s", serviceName))

	if err != nil {
		return nil, err
	}

	var spans [][]model.SpanModel
	err = json.Unmarshal([]byte(traces), &spans)
	if err != nil {
		return nil, err
	}

	return spans, nil

}
