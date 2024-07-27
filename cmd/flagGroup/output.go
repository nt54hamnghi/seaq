package flagGroup

import (
	"errors"
	"io"
	"os"

	"github.com/nt54hamnghi/hiku/pkg/util"
	"github.com/spf13/cobra"
)

type Output struct {
	File  string
	Force bool
}

func (o *Output) Writer() (io.WriteCloser, error) {
	if o.File == "" {
		return os.Stdout, nil
	}

	if o.Force {
		return util.NewOverwriteWriter(o.File)
	} else {
		return util.NewFailExistingWriter(o.File)
	}
}

func (o *Output) Validate(cmd *cobra.Command, _ []string) error {
	outputFileSet := cmd.Flags().Changed("output")
	forceSet := cmd.Flags().Changed("force")

	if forceSet && (!outputFileSet || o.File == "") {
		return errors.New("--force can only be used with --output")
	}

	return nil
}

func (o *Output) Init(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.File, "output", "o", "", "output file")
	cmd.Flags().BoolVarP(&o.Force, "force", "f", false, "overwrite existing file")
}
