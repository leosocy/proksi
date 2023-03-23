// Copyright (c) 2019 leosocy, leosocy@gmail.com
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
type rootCmd struct {
	configFile string

	cmd *cobra.Command
}

// Command returns root command
func (cc *rootCmd) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proksi",
		Short: "Proksi is an intelligent proxy server.",
		Long: `Proksi is an intelligent proxy server.
It provides durable, real-time, high-quality proxies as a middleman or datasource server.`,
		PersistentPreRunE: cc.initConfig,
	}

	cmd.PersistentFlags().StringVar(&cc.configFile, "config", "", "config file (default is $HOME/.proksi/config.yaml)")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	cc.cmd = cmd
	return cc.cmd
}

// initConfig initializes config
func (cc *rootCmd) initConfig(cmd *cobra.Command, args []string) error {
	fmt.Println(cc.configFile)
	return nil
}

var (
	rc *cobra.Command
)

func init() {
	rc = (&rootCmd{}).Command()
	serverCmd.AddCommand((&serverMitmCmd{}).Command())
	rc.AddCommand(serverCmd, versionCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rc.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
