/*
Copyright © 2024 Nghi Nguyen
*/
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hoc",
	Short: "A cli tool to make learning more fun",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := readStdin()

		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}

		fmt.Println(input)

		// fmt.Println(string(input))
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

func init() {}

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
