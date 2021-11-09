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
	"strconv"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/client/pkg/kn/flags"

	"knative.dev/client/pkg/kn/commands"
)

type configUpdateFlags struct {
	debug bool
}

func (c *configUpdateFlags) addFlags(cmd *cobra.Command) {
	flags.AddBothBoolFlags(cmd.Flags(), &c.debug, "debug", "d", false, "set tracing debug mode.")
}

// NewUpdateCommand implements 'kn trace config update' command
func NewUpdateCommand(p *commands.KnParams) *cobra.Command {
	var updateflags configUpdateFlags

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update tracing configuration parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := flags.ReconcileBoolFlags(cmd.Flags())
			if err != nil {
				return err
			}

			cfg, err := p.RestConfig()
			if err != nil {
				return fmt.Errorf("failed to update tracing configuration: %w", err)
			}

			client, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return fmt.Errorf("failed to update tracing configuration: %w", err)
			}

			// TODO: alternate eventing installation namespace
			cm, err := client.CoreV1().ConfigMaps("knative-eventing").Get(cmd.Context(), "config-tracing", metav1.GetOptions{})
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("\"failed to update tracing configuration: %w", err)
				}

				// knative eventing hasn't been installed properly.
				fmt.Println("⚠️ missing config-tracing in the knative-eventing namespace which is an indicator that Knative Eventing hasn't been properly installed (recovering)")
				cm = &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: "config-tracing",
					},
					Data: map[string]string{},
				}
			}

			setDebug := cmd.Flags().Changed("debug") || cmd.Flags().Changed("no-debug")
			updated := false

			if setDebug {
				debugStr := strconv.FormatBool(updateflags.debug)

				debug, ok := cm.Data["debug"]
				if !ok || debug != debugStr {
					cm.Data["debug"] = debugStr
					updated = true
				}
			}

			if updated {
				_, err = client.CoreV1().ConfigMaps("knative-eventing").Update(cmd.Context(), cm, metav1.UpdateOptions{})
				if err != nil {
					return err
				}

				fmt.Println("✔️tracing configuration successfully modified")
				return nil
			}

			fmt.Println("✔️tracing configuration unchanged")
			return nil
		},
	}

	updateflags.addFlags(cmd)

	return cmd
}
