/*
Copyright © 2024 Nghi Nguyen
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/nt54hamnghi/hiku/cmd/scrape"
	"github.com/nt54hamnghi/hiku/pkg/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hiku",
	Short: "A cli tool to make learning more fun",
	Args:  cobra.NoArgs,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		noStream, err := cmd.Flags().GetBool("no-stream")
		if err != nil {
			return err
		}

		input, err := readStdin()

		if err != nil && err.Error() == "interactive input is not supported" {
			cmd.Help()
			return nil
		}

		ctx := context.Background()
		if noStream {
			openai.CreateCompletion(ctx, openai.PrimingPrompt, input)
		} else {
			openai.CreateCompletionStream(ctx, openai.PrimingPrompt, input)
		}

		return nil
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

func addCommandPallete() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scrape.ScrapeCmd)
}

func init() {
	// init viper config and register it with cobra
	cobra.OnFinalize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hiku.yaml)")

	rootCmd.Flags().Bool("no-stream", false, "disable streaming mode")
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	addCommandPallete()
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
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("hiku")
		// Path to look for the config file in
		// The order of paths listed is the order in which they will be searched
		viper.AddConfigPath("/etc/hiku")
		viper.AddConfigPath("$HOME/.config/hiku")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
