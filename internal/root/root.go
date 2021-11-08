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

package root

import (
	"github.com/spf13/cobra"
	"knative.dev/kn-plugin-trace/internal/commands/config"
	"knative.dev/kn-plugin-trace/internal/commands/show"

	clientcmds "knative.dev/client/pkg/kn/commands"

	"knative.dev/kn-plugin-trace/internal/commands"
)

// NewRootCommand represents the plugin's entrypoint
func NewRootCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "kn-trace",
		Short: "Manage traces",
		Long:  "Manage traces",
	}

	p := &clientcmds.KnParams{}
	p.Initialize()

	rootCmd.AddCommand(config.NewConfigCommand(p))

	rootCmd.AddCommand(show.NewShowCommand(p))
	rootCmd.AddCommand(commands.NewVersionCommand())

	return rootCmd
}
