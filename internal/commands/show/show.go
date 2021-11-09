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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"knative.dev/kn-plugin-trace/pkg/zipkin"

	"knative.dev/client/pkg/kn/commands"

	"knative.dev/kn-plugin-trace/pkg/config"
)

type showFlags struct {
	follow bool
}

func (c *showFlags) addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&c.follow, "follow", "f", false, "whether the traces should be streamed")
}

// NewShowCommand is the command for showing traces
func NewShowCommand(p *commands.KnParams) *cobra.Command {
	var showflags showFlags

	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show traces",

		RunE: func(cmd *cobra.Command, args []string) error {
			restcfg, err := p.RestConfig()
			if err != nil {
				return err
			}

			// Read Tracing configuration

			kubeclient, err := kubernetes.NewForConfig(restcfg)
			if err != nil {
				return err
			}

			cfg, err := config.Load(cmd.Context(), kubeclient)
			if err != nil {
				return err
			}

			if err := config.Validate(cfg); err != nil {
				return err
			}

			// Create connection to Zipkin
			connection, err := zipkin.Connect(cfg.ZipkinEndpoint, restcfg)
			if err != nil {
				return err
			}

			since := time.UnixMilli(0)
			for {
				now := time.Now()

				err := showSpans(connection, now, since)
				if err != nil {
					return err
				}

				if !showflags.follow {
					return nil
				}

				time.Sleep(1 * time.Second)
				since = now
			}

			return nil

		},
	}

	showflags.addFlags(showCmd)

	return showCmd
}

func showSpans(connection *zipkin.Connection, now time.Time, since time.Time) error {
	endTs := now
	lookback := endTs.Sub(since).Milliseconds()

	// Get all traces
	services, err := connection.Services()
	if err != nil {
		return err
	}

	for _, svc := range services {
		spans, err := connection.Spans(svc, endTs.UnixMilli(), lookback)

		if err != nil {
			return err
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

}
