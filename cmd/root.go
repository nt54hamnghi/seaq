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
	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var patternName string

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

		cfg, err := pattern.WithName(patternName)
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		prompt, err := cfg.GetPrompt()
		if err != nil {
			cmd.SilenceUsage = true
			return err
		}

		input, err := readStdin()
		if err != nil && err.Error() == "interactive input is not supported" {
			cmd.Help()
			return nil
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

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/hiku.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "verbose output")
	rootCmd.Flags().Bool("no-stream", false, "disable streaming mode")
	rootCmd.Flags().StringVarP(&patternName, "pattern", "p", "", "pattern to use for completion")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	addCommandPallete()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".hiku" (without extension).
		viper.SetConfigName("hiku")
		viper.SetConfigType("yaml")

		// Path to look for the config file in
		// The order of paths listed is the order in which they will be searched
		viper.AddConfigPath("/etc/hiku")
		viper.AddConfigPath(filepath.Join(home, ".config/hiku"))
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
