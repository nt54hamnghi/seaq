/*
Copyright © 2024 Nghi Nguyen
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nt54hamnghi/hiku/cmd/pattern"
	"github.com/nt54hamnghi/hiku/cmd/scrape"
	"github.com/nt54hamnghi/hiku/pkg/openai"
	"github.com/spf13/cobra"
)

var configFile string
var patternName string
var patternRepo string
var verbose bool
var Hiku *HikuConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "hiku",
	Short:        "A cli tool to make learning more fun",
	Version:      "0.1.0",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// get the value of the no-stream flag
		noStream, err := cmd.Flags().GetBool("no-stream")
		if err != nil {
			return err
		}

		input, err := readStdin()
		if err != nil && err.Error() == "interactive input is not supported" {
			cmd.Help()
			return nil
		}

		prompt, err := Hiku.GetPrompt()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		if input == "" {
			return fmt.Errorf("piped input is empty")
		}

		ctx := context.Background()
		if noStream {
			openai.CreateCompletion(ctx, prompt, input)
		} else {
			openai.CreateCompletionStream(ctx, prompt, input)
		}

		return nil
	},
}

// Read from stdin
// This function allow piping input to the command.
// It'll call the fallback function if stdin is provided interactively.
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
		return "", fmt.Errorf("interactive input is not supported")
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

func addCommandPallete() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scrape.ScrapeCmd)
	rootCmd.AddCommand(pattern.PatternCmd)
}

func init() {
	// init viper config and register it with cobra
	cobra.OnInitialize(initConfig)

	// settings
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// flags definition
	// persistent flags are global and available to all commands
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.config/hiku.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "verbose output")

	// local flags are only available to the root command
	rootCmd.Flags().Bool("no-stream", false, "disable streaming mode")
	rootCmd.Flags().StringVarP(&patternName, "pattern", "p", "", "pattern to use for completion")
	rootCmd.Flags().StringVarP(&patternRepo, "repo", "r", "", "path to the pattern repository")

	// add subcommands
	addCommandPallete()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	Hiku = New()
	if configFile != "" {
		// Use config file from the flag.
		Hiku.SetConfigFile(configFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".hiku" (without extension).
		Hiku.SetConfigName("hiku")
		Hiku.SetConfigType("yaml")

		// Path to look for the config file in
		// The order of paths listed is the order in which they will be searched
		Hiku.AddConfigPath("/etc/hiku")
		Hiku.AddConfigPath(filepath.Join(home, ".config/hiku"))
		Hiku.AddConfigPath(".")
	}

	// bind flags to viper
	Hiku.BindPFlag("pattern.name", rootCmd.Flags().Lookup("pattern"))
	Hiku.BindPFlag("pattern.repo", rootCmd.Flags().Lookup("repo"))
	Hiku.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := Hiku.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", Hiku.ConfigFileUsed())
	}
}
