/*
 * Copyright 2020 ZUP IT SERVICOS EM TECNOLOGIA E INOVACAO SA
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"github.com/spf13/cobra"
)

// NewAddCmd create a new add instance.
func NewAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "add SUBCOMMAND",
		Short:     "Add repositories and workspaces",
		Long:      "Add a new repository of formulas or a new workspace",
		Example:   "rit add repo",
		ValidArgs: []string{"repo", "workspace"},
		Args:      cobra.OnlyValidArgs,
	}
}
