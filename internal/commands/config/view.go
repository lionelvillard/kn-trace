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

package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"knative.dev/kn-plugin-trace/internal/output"
	"knative.dev/kn-plugin-trace/pkg/zipkin"

	"knative.dev/client/pkg/kn/commands"
	"knative.dev/kn-plugin-trace/pkg/config"
)

// NewViewCommand implements 'kn trace config info' command
func NewViewCommand(p *commands.KnParams) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View the current tracing configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			restcfg, err := p.RestConfig()
			if err != nil {
				return err
			}

			kubeclient, err := kubernetes.NewForConfig(restcfg)
			if err != nil {
				return err
			}

			cfg, err := config.Load(cmd.Context(), kubeclient)
			if err != nil {
				return err
			}

			if cfg.Backend == "zipkin" || cfg.Backend == "none" {
				output.Checkmark()
			} else {
				output.Error()
			}

			fmt.Printf("backend: %s\n", cfg.Backend)

			if cfg.Backend == "zipkin" {
				if cfg.ZipkinEndpoint == "" {
					output.Error()
				} else {
					output.Checkmark()
					fmt.Printf("zipkinEndpoint: %s\n", cfg.ZipkinEndpoint)

					if _, err := zipkin.Connect(cfg.ZipkinEndpoint, restcfg); err == nil {
						output.Checkmark()
						fmt.Println("Reachable")
					} else {
						output.Error()
						fmt.Println("Unreachable")
					}
				}
			}

			if cfg.Debug == false {
				output.Warning()
				fmt.Printf("debug: %t (only some traces will be displayed when running kn trace show)\n", cfg.Debug)
			} else {
				output.Checkmark()
				fmt.Printf("debug: %t\n", cfg.Debug)
			}

			output.Checkmark()
			fmt.Printf("sample-rate: %f\n", cfg.SampleRate)
			return nil
		},
	}

	return cmd
}
