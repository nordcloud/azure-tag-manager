package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pkg/errors"

	"github.com/nordcloud/azure-tag-manager/internal/azure"
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
)

const (
	usageResourceGroup = "Specifies resource group"
)

var (
	verboseEnabled bool
	resourceGroup  string
)

func init() {
	rootCmd.AddCommand(checkCommand)
	checkCommand.Flags().StringVarP(&resourceGroup, "rg", "r", "", usageResourceGroup)
	checkCommand.MarkFlagRequired("rg")
}

var checkCommand = &cobra.Command{
	Use:   "check",
	Short: "Do sanity checks on a resource group (NOT FULLY IMPLEMENTED YET)",
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.NewFromFile()
		if err != nil {
			return errors.Wrap(err, "could not create session")
		}

		scanner := azure.NewResourceGroupScanner(sess)
		res, err := scanner.GetResourcesByResourceGroup(resourceGroup)
		if err != nil {
			return errors.Wrap(err, "could not get resources by group")
		}

		checker := azure.TagChecker{
			Session: sess,
		}
		fmt.Printf("Checking same tag with different values in [%s]\n", resourceGroup)
		nonc := checker.CheckSameTagDifferentValue(res)
		for tag, nonrList := range nonc {
			fmt.Printf("Noncompliant tag [%s]\n", tag)
			for _, nonr := range nonrList {
				fmt.Printf("Seen [%s] = [%s] in [%s]\n", tag, nonr.Value, nonr.Resource.ID)
			}
		}

		if len(nonc) == 0 {
			fmt.Printf("💪  Resource group [%s] has no tags with different values\n", resourceGroup)
		}

		return nil
	}}
