/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"zssh/zsshlib"

	"github.com/openziti/ziti/common/enrollment"
	"github.com/openziti/ziti/ziti/cmd/common"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const ExpectedServiceAndExeName = "zssh"

var (
	flags = zsshlib.SshFlags{}
)

var rootCmd = &cobra.Command{
	Use:   fmt.Sprintf("%s <remoteUsername>@<targetIdentity>", flags.ServiceName),
	Short: "Z(iti)ssh, Carb-loaded ssh performs faster and stronger than ssh",
	Long:  "Z(iti)ssh is a version of ssh that utilizes a ziti network to provide a faster and more secure remote connection. A ziti connection must be established before use",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("You need to specify at least one positional argument")
			os.Exit(1)
		}

		username, targetIdentity := flags.GetUserAndIdentity(args[0])

		cmdArgs := args[1:]
		token := ""
		var err error
		if flags.OIDC.Mode {
			token, err = zsshlib.OIDCFlow(context.Background(), flags)
			if err != nil {
				logrus.Fatalf("error performing OIDC flow: %v", err)
			}
		}
		sshConn := zsshlib.EstablishClient(flags, username, targetIdentity, token)
		defer func() { _ = sshConn.Close() }()
		err = zsshlib.RemoteShell(sshConn, cmdArgs)
		if err != nil {
			logrus.Fatalf("error opening remote shell: %v", err)
		}
	},
}

func init() {
	flags.InitFlags(rootCmd, ExpectedServiceAndExeName)
	flags.OIDCFlags(rootCmd, ExpectedServiceAndExeName)
}

// AuthCmd holds the required data for the init cmd
type AuthCmd struct {
	common.OptionsProvider
}

func NewAuthCmd(p common.OptionsProvider) *cobra.Command {
	cmd := &AuthCmd{OptionsProvider: p}

	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Test OIDC auth flow",
		Long:  `Test authentication against IdP to get OIDC ID token.`,
		Args:  cobra.NoArgs,
		RunE:  cmd.Run,
	}
	flags.OIDCFlags(authCmd, ExpectedServiceAndExeName)
	return authCmd
}

func (cmd *AuthCmd) Run(_ *cobra.Command, _ []string) error {
	_, err := zsshlib.OIDCFlow(context.Background(), flags)
	return err
}

func main() {
	p := common.NewOptionsProvider(os.Stdout, os.Stderr)
	rootCmd.AddCommand(enrollment.NewEnrollCommand(p))
	// leave out for now // rootCmd.AddCommand(NewAuthCmd(p))
	e := rootCmd.Execute()
	if e != nil {
		logrus.Error(e)
	}
}
