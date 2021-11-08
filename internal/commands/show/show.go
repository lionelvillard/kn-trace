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

package show

import (
	"fmt"
	"net/url"
	"regexp"

	model "github.com/openzipkin/zipkin-go/model"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"knative.dev/kn-plugin-trace/pkg/zipkin"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/pkg/webhook/json"

	"knative.dev/kn-plugin-trace/pkg/config"
	"knative.dev/kn-plugin-trace/pkg/proxy"
)

type showFlags struct {
}

func (c *showFlags) addFlags(cmd *cobra.Command) {

}

// NewShowCommand is the command for showing traces
func NewShowCommand(p *commands.KnParams) *cobra.Command {
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show traces",

		RunE: func(cmd *cobra.Command, args []string) error {
			restcfg, err := p.RestConfig()
			if err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			// Read Tracing configuration

			kubeclient, err := kubernetes.NewForConfig(restcfg)
			if err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			cfg, err := config.Load(cmd.Context(), kubeclient)
			if err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			if err := config.Validate(cfg); err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			// Create Service proxy

			pclient, err := proxy.New(restcfg)
			if err != nil {
				return err
			}

			// Get all traces

			// TODO: support externally accessible traces

			url, err := url.Parse(cfg.ZipkinEndpoint)
			if err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			parts := regexp.MustCompile("[.:]").Split(url.Host, -1)
			fmt.Println(parts)

			if len(parts) < 2 {
				return fmt.Errorf("failed to show traces: invalid Zipkin endpoint %s", cfg.ZipkinEndpoint)
			}

			svcName := parts[0]
			svcNamespace := parts[1]

			resp, err := pclient.Get(svcName, svcNamespace, "api/v2/services")
			if err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			var services zipkin.ServicesResponse
			err = json.Unmarshal([]byte(resp), &services)
			if err != nil {
				return fmt.Errorf("failed to show traces: %w", err)
			}

			for _, svc := range services {
				traces, err := pclient.Get(svcName, svcNamespace, fmt.Sprintf("api/v2/traces?serviceName=%s", svc))

				if err != nil {
					// TODO: ignore?
					return fmt.Errorf("failed to show traces: %w", err)
				}

				var spans [][]model.SpanModel
				err = json.Unmarshal([]byte(traces), &spans)
				if err != nil {
					return fmt.Errorf("failed to show traces: %w", err)
				}

				for _, span1 := range spans {

					for _, span := range span1 {
						// Just show cloudevents
						if span.Name == "cloudevents.client" {
							fmt.Printf("%s %s %s\n", span.Tags["cloudevents.source"], span.Tags["cloudevents.id"], span.Tags["cloudevents.type"])
						}
					}
				}
			}

			return nil
		},
	}

	return showCmd
}
