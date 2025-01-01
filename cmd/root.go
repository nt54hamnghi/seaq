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
	"github.com/nt54hamnghi/hiku/cmd/flaggroup"
	"github.com/nt54hamnghi/hiku/cmd/model"
	"github.com/nt54hamnghi/hiku/cmd/pattern"
	"github.com/nt54hamnghi/hiku/pkg/llm"
	"github.com/nt54hamnghi/hiku/pkg/util"

	"github.com/spf13/cobra"
)

// region: --- flags

var (
	configFile  string
	hint        string
	modelName   string
	noStream    bool
	output      flaggroup.Output
	patternName string
	patternRepo string
	verbose     bool

	common = &cobra.Group{
		Title: "Common Commands:",
		ID:    "common",
	}

	management = &cobra.Group{
		Title: "Management Commands:",
		ID:    "management",
	}
)

// endregion: --- flags

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "hiku",
	Short:        "A cli tool to make learning more fun",
	Version:      "0.1.5",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	PreRunE:      flaggroup.ValidateGroups(&output),
	RunE: func(cmd *cobra.Command, args []string) error { // nolint: revive
		hiku := config.Hiku

		input, err := util.ReadPipedStdin()
		if err != nil {
			if errors.Is(err, util.ErrInteractiveInput) {
				_ = cmd.Help()
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
		msgs := llm.PrepareMessages(prompt, input, hint)

		if noStream {
			return llm.CreateCompletion(ctx, model, dest, msgs)
		}
		return llm.CreateStreamCompletion(ctx, model, dest, msgs)
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

	// flags setting
	flags := rootCmd.Flags()
	pFlags := rootCmd.PersistentFlags()

	// flags definition
	// persistent flags are global and available to all commands
	pFlags.StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.config/hiku.yaml)")
	pFlags.BoolVarP(&verbose, "verbose", "V", false, "verbose output")

	// local flags are only available to the root command
	flags.SortFlags = false
	flags.StringVarP(&hint, "hint", "i", "", "optional context to guide the LLM's focus")
	flags.BoolVar(&noStream, "no-stream", false, "disable streaming mode")
	flags.StringVarP(&patternRepo, "repo", "r", "", "path to the pattern repository")
	flags.StringVarP(&patternName, "pattern", "p", "", "pattern to use")
	flags.StringVarP(&modelName, "model", "m", "", "model to use")

	// register completion function
	_ = rootCmd.RegisterFlagCompletionFunc("pattern", pattern.CompletePatternArgs)
	_ = rootCmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)

	// flag groups
	flaggroup.InitGroups(rootCmd, &output)

	// assign commands to groups
	// https://github.com/spf13/cobra/blob/main/site/content/user_guide.md#grouping-commands-in-help
	chat.ChatCmd.GroupID = "common"
	fetch.FetchCmd.GroupID = "common"
	pattern.PatternCmd.GroupID = "management"
	model.ModelCmd.GroupID = "management"

	// add subcommands
	rootCmd.AddCommand(
		chat.ChatCmd,
		fetch.FetchCmd,
		pattern.PatternCmd,
		model.ModelCmd,
	)

	// add groups
	rootCmd.AddGroup(
		common,
		management,
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
	flags := rootCmd.Flags()
	if err := hiku.BindPFlag("pattern.name", flags.Lookup("pattern")); err != nil {
		return err
	}
	if err := hiku.BindPFlag("pattern.repo", flags.Lookup("repo")); err != nil {
		return err
	}
	if err := hiku.BindPFlag("model.name", flags.Lookup("model")); err != nil {
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
