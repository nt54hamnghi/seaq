package flagGroup

import "github.com/spf13/cobra"

type PreRunE func(cmd *cobra.Command, args []string) error

type FlagGroup interface {
	Init(cmd *cobra.Command)
	Validate(cmd *cobra.Command, args []string) error
}

func ValidateGroups(groups ...FlagGroup) PreRunE {
	return func(cmd *cobra.Command, args []string) error {
		for _, group := range groups {
			if err := group.Validate(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func InitGroups(cmd *cobra.Command, groups ...FlagGroup) {
	for _, group := range groups {
		group.Init(cmd)
	}
}
