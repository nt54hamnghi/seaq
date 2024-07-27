/*
Copyright © 2024 Nghi Nguyen
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nt54hamnghi/hiku/cmd/config"
	"github.com/nt54hamnghi/hiku/cmd/fetch"
	"github.com/nt54hamnghi/hiku/cmd/flagGroup"
	"github.com/nt54hamnghi/hiku/cmd/model"
	"github.com/nt54hamnghi/hiku/cmd/pattern"
	"github.com/nt54hamnghi/hiku/pkg/llm"

	"github.com/spf13/cobra"
)

// region: --- errors

var errInteractiveInput = errors.New("interactive input is not supported")

// endregion: --- errors

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
	PreRunE:      output.Validate,
	RunE: func(cmd *cobra.Command, args []string) error {
		// read from stdin if it's piped
		input, err := readStdin()
		if err != nil && errors.Is(err, errInteractiveInput) {
			cmd.Help()
			return nil
		}

		// construct the prompt from pattern and scraped content
		prompt, err := config.Hiku.GetPrompt()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		// check if input is empty
		if input == "" {
			return errors.New("piped input is empty")
		}

		if verbose {
			fmt.Println("Using model:", config.Hiku.Model())
		}

		// construct the model
		model, err := llm.New(config.Hiku.Model())
		if err != nil {
			return err
		}

		dest, err := output.Writer()
		if err != nil {
			return err
		}
		defer dest.Close()

		// run the completion
		msgs := llm.PrepareMessages(prompt, input)
		if noStream {
			return llm.CreateCompletion(context.Background(), model, msgs, dest)
		} else {
			return llm.CreateStreamCompletion(context.Background(), model, msgs, dest)
		}
	},
}

// readStdin reads from stdin if it's piped, otherwise it returns an error.
func readStdin() (string, error) {
	// get stat of stdin file descriptor
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// Check if input is piped or interactive
	//
	// `info.Mode()` returns the file mode bits
	// `os.ModeCharDevice` is a file mode bit for a character device
	// when input is piped, `info.Mode()` will NOT have `os.ModeCharDevice` set
	// -> a bitwise AND with `os.ModeCharDevice` will return 0
	// when input is interactive, `info.Mode()` will have `os.ModeCharDevice` set
	// -> a bitwise AND with `os.ModeCharDevice` will return a non-zero value
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", errInteractiveInput
	}

	// read from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	return string(input), nil
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
	cobra.OnInitialize(initConfig)

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
	output.Init(rootCmd)

	// register completion function
	rootCmd.RegisterFlagCompletionFunc("pattern", pattern.CompletePatternArgs)
	rootCmd.RegisterFlagCompletionFunc("model", model.CompleteModelArgs)

	// add subcommands
	rootCmd.AddCommand(
		fetch.FetchCmd,
		pattern.PatternCmd,
		model.ModelCmd,
	)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	config.Hiku = config.New()
	if configFile != "" {
		// Use config file from the flag.
		config.Hiku.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".hiku" (without extension).
		config.Hiku.SetConfigName("hiku")
		config.Hiku.SetConfigType("yaml")

		// Path to look for the config file in
		// The order of paths listed is the order in which they will be searched
		config.Hiku.AddConfigPath("/etc/hiku")
		config.Hiku.AddConfigPath(filepath.Join(home, ".config/hiku"))
		config.Hiku.AddConfigPath(".")
	}

	// bind flags to viper

	// -- pattern
	config.Hiku.BindPFlag("pattern.name", rootCmd.Flags().Lookup("pattern"))
	config.Hiku.BindPFlag("pattern.repo", rootCmd.Flags().Lookup("repo"))

	// -- model
	config.Hiku.BindPFlag("model.name", rootCmd.Flags().Lookup("model"))

	config.Hiku.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := config.Hiku.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", config.Hiku.ConfigFileUsed())
	}
}
