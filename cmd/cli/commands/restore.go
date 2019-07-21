package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/nordcloud/azure-tag-manager/internal/azure"
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
)

const (
	usageRestoreFile = "Specify the location of the restore file"
)

var (
	restoreFile string
)

func init() {
	rootCmd.AddCommand(restoreCommand)
	restoreCommand.Flags().StringVarP(&restoreFile, "file", "f", "", usageRestoreFile)
	restoreCommand.MarkFlagRequired("file")
}

var restoreCommand = &cobra.Command{
	Use:   "restore",
	Short: "Restore previous tags from a file backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.NewFromFile()
		if err != nil {
			return errors.Wrap(err, "could not create session")
		}

		fmt.Printf("Restoring tags from: [%s]\n", restoreFile)

		restorer := azure.NewRestorerFromFile(restoreFile, sess, false)
		err = restorer.Restore()

		if err != nil {
			return errors.Wrap(err, "could not restore backup")
		}
		return nil
	},
}
