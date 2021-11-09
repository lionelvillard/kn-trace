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
	"knative.dev/kn-plugin-trace/pkg/config"
	"knative.dev/kn-plugin-trace/pkg/setup"

	"knative.dev/client/pkg/kn/commands"
)

type configEnableFlags struct {
	template string
}

func (c *configEnableFlags) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.template, "template", "t", "zipkin", "tracing configuration template")
	cobra.MarkFlagRequired(cmd.Flags(), "template")
}

// NewEnableCommand implements 'kn trace config enable' command
func NewEnableCommand(p *commands.KnParams) *cobra.Command {
	var enableFlags configEnableFlags

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable tracing",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if tracing is already enabled.
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

			if cfg.Backend == "" || cfg.Backend == "none" {
				switch enableFlags.template {
				case "zipkin":
					err := setup.Zipkin(cmd.Context(), p)
					if err != nil {
						return err
					}
					fmt.Println("tracing enabled")
					return nil

				default:
					return fmt.Errorf("invalid template %s", enableFlags.template)
				}
			}

			output.Checkmark()
			fmt.Println("tracing is already enabled")
			return nil
		},
	}

	enableFlags.addFlags(cmd)

	return cmd
}
