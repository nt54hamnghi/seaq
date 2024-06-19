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

	"github.com/nt54hamnghi/hiku/cmd/scrape"
	"github.com/nt54hamnghi/hiku/pkg/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var pattern string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hiku",
	Short: "A cli tool to make learning more fun",
	Args:  cobra.NoArgs,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		// get the value of the no-stream flag
		noStream, err := cmd.Flags().GetBool("no-stream")
		if err != nil {
			return err
		}

		cfg, err := loadPatternConfig()
		if err != nil {
			return err
		}

		prompt, err := cfg.getPrompt()
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
			openai.CreateCompletion(ctx, prompt, input)
		} else {
			openai.CreateCompletionStream(ctx, prompt, input)
		}

		return nil
	},
}

type patternConfig struct {
	repo    string
	pattern string
}

func loadPatternConfig() (cfg patternConfig, err error) {
	// use viper to read the config file
	// get the patterns repo path
	cfg.repo = viper.GetString("patterns.repo")
	if cfg.repo == "" {
		return cfg, fmt.Errorf("patterns.repo is not set in the config file")
	}

	// get the default prompt
	cfg.pattern = pattern // set to the pattern flag value
	if cfg.pattern == "" {
		cfg.pattern = viper.GetString("patterns.default")
	}

	if cfg.pattern == "" {
		return cfg, fmt.Errorf("no pattern provided")
	}

	return
}

func (cfg patternConfig) getPrompt() (string, error) {
	// read the pattern
	prompt, err := os.ReadFile(filepath.Join(cfg.repo, cfg.pattern, "system.md"))
	if err != nil {
		return "", err
	}

	return string(prompt), nil
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/hiku.yaml)")
	rootCmd.Flags().Bool("no-stream", false, "disable streaming mode")
	rootCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "pattern to use for completion")
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
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
