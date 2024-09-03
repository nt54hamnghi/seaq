/*
Copyright © 2024 Nghi Nguyen
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/nt54hamnghi/hiku/cmd/chat"
	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/nt54hamnghi/hiku/cmd/fetch"
	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
	"github.com/nt54hamnghi/hiku/cmd/model"
	"github.com/nt54hamnghi/hiku/cmd/pattern"
	"github.com/nt54hamnghi/hiku/pkg/llm"
	"github.com/nt54hamnghi/hiku/pkg/util"

	"github.com/spf13/cobra"
)

// region: --- flags

var (
	configFile  string
	patternName string
	patternRepo string
	modelName   string
	noStream    bool
	verbose     bool
	output      flagGroup.Output
)

// endregion: --- flags

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "hiku",
	Short:        "A cli tool to make learning more fun",
	Version:      "0.1.0",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	PreRunE:      flagGroup.ValidateGroups(&output),
	RunE: func(cmd *cobra.Command, args []string) error {
		hiku := config.Hiku

		input, err := util.ReadPipedStdin()
		if err != nil {
			if errors.Is(err, util.ErrInteractiveInput) {
				cmd.Help()
				return nil
			}
			return err
		}

		// construct the prompt from pattern and scraped content
		prompt, err := hiku.GetPrompt()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		if verbose {
			fmt.Println("Using model:", hiku.Model())
		}

		// construct the model
		model, err := llm.New(hiku.Model())
		if err != nil {
			return err
		}

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		// run the completion
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		msgs := llm.PrepareMessages(prompt, string(input))

		if noStream {
			return llm.CreateCompletion(ctx, model, dest, msgs)
		} else {
			return llm.CreateStreamCompletion(ctx, model, dest, msgs)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// init viper config and register it with cobra
	cobra.OnInitialize(func() {
		cobra.CheckErr(initConfig())
	})

	// flags definition
	// persistent flags are global and available to all commands
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.config/hiku.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "verbose output")

	// local flags are only available to the root command
	rootCmd.Flags().BoolVar(&noStream, "no-stream", false, "disable streaming mode")
	rootCmd.Flags().StringVarP(&patternRepo, "repo", "r", "", "path to the pattern repository")
	rootCmd.Flags().StringVarP(&patternName, "pattern", "p", "", "pattern to use")
	rootCmd.Flags().StringVarP(&modelName, "model", "m", "", "model to use")

	// flag groups
	flagGroup.InitGroups(rootCmd, &output)

	// register completion function
	rootCmd.RegisterFlagCompletionFunc("pattern", pattern.CompletePatternArgs)
	rootCmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)

	// add subcommands
	rootCmd.AddCommand(
		chat.ChatCmd,
		fetch.FetchCmd,
		pattern.PatternCmd,
		model.ModelCmd,
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// bind the global HikuConfig to a local variable
	hiku := config.Hiku

	// set the config file if provided otherwise search for it
	if configFile != "" {
		hiku.SetConfigFile(configFile)
	} else if err := hiku.SearchConfigFile(); err != nil {
		return err
	}

	// bind flags to viper
	flgs := rootCmd.Flags()
	if err := hiku.BindPFlag("pattern.name", flgs.Lookup("pattern")); err != nil {
		return err
	}
	if err := hiku.BindPFlag("pattern.repo", flgs.Lookup("repo")); err != nil {
		return err
	}
	if err := hiku.BindPFlag("model.name", flgs.Lookup("model")); err != nil {
		return err
	}

	// read in the config file and environment variables
	hiku.AutomaticEnv()

	// If a config file is found, read it in.
	if err := hiku.ReadInConfig(); err != nil {
		return err
	}
	if verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", hiku.ConfigFileUsed())
	}

	return nil
}
