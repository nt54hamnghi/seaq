package compose

import "github.com/spf13/cobra"

type (
	Run  = func(cmd *cobra.Command, args []string)
	RunE = func(cmd *cobra.Command, args []string) error
)

// Sequence runs a run command in order, one after the other.
func Sequence(runs ...Run) Run {
	return func(cmd *cobra.Command, args []string) {
		for _, run := range runs {
			run(cmd, args)
		}
	}
}

// SequenceE runs a runE command in order, one after the other
// and return on the first error.
func SequenceE(runs ...RunE) RunE {
	return func(cmd *cobra.Command, args []string) error {
		for _, run := range runs {
			if err := run(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
