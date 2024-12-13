package flaggroup

import (
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
	"github.com/spf13/cobra"
)

type Interval struct {
	Start timestamp.Timestamp
	End   timestamp.Timestamp
}

func (i *Interval) Validate(cmd *cobra.Command, args []string) error { // nolint: revive
	return nil
}

func (i *Interval) Init(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.VarP(&i.Start, "start", "s", "start time")
	flags.VarP(&i.End, "end", "e", "end time")
}
