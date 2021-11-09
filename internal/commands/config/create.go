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
	"knative.dev/kn-plugin-trace/pkg/setup"

	"knative.dev/client/pkg/kn/commands"
)

const (
	KnToolsNamespace = "kntools"
)

type configCreateFlags struct {
	template string
}

func (c *configCreateFlags) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&c.template, "template", "t", "zipkin", "tracing configuration template")
	cobra.MarkFlagRequired(cmd.Flags(), "template")
}

// NewCreateCommand implements 'kn trace config update' command
func NewCreateCommand(p *commands.KnParams) *cobra.Command {
	var createFlags configCreateFlags

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create tracing configuration from template",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch createFlags.template {
			case "zipkin":
				return setup.Zipkin(cmd.Context(), p)
			default:
				return fmt.Errorf("invalid template %s", createFlags.template)
			}
		},
	}

	createFlags.addFlags(cmd)

	return cmd
}
