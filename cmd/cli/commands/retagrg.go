package commands

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cleanTags bool

func init() {
	rootCmd.AddCommand(resourceGroupTagCommand)
	resourceGroupTagCommand.Flags().StringVarP(&resourceGroup, "rg", "r", "", usageResourceGroup)
	resourceGroupTagCommand.Flags().BoolVar(&cleanTags, "cleantags", false, "Clean all tags before adding")
	resourceGroupTagCommand.MarkFlagRequired("rg")
	resourceGroupTagCommand.Flags().BoolVar(&dryRunEnabled, "dry", false, usageDryRun)

}

var resourceGroupTagCommand = &cobra.Command{
	Use:   "retagrg",
	Short: "Retag resources in a rg based on tags on rgs",
	Long:  "Takes tags form a given resource group and applies them to all of the resources in the resource group. If any existing tags are already there, the new ones with be appended.",
	RunE: func(cmd *cobra.Command, args []string) error {

		sess, err := session.NewFromFile()
		if err != nil {
			return errors.Wrap(err, "Could not create session")
		}
		scanner := azure.NewResourceGroupScanner(sess)

		rgTags, err := scanner.GetResourceGroupTags(resourceGroup)

		if err != nil {
			return errors.Wrap(err, "Can't get tags")
		}

		resources := scanner.ScanResourceGroup(resourceGroup)

		var actions []rules.ActionItem

		if cleanTags {
			actions = append(actions, rules.ActionItem{"type": "cleanTags"})
		}

		for key, tag := range rgTags {
			actions = append(actions, rules.ActionItem{"type": "addTag", "tag": key, "value": *tag})
		}

		rules := rules.TagRules{Rules: []rules.Rule{
			rules.Rule{Name: "name", Conditions: []rules.ConditionItem{
				rules.ConditionItem{"type": "rgEqual", "resourceGroup": resourceGroup},
			},
				Actions: actions,
			},
		}}

		tagger := azure.NewTagger(rules, sess)
		if dryRunEnabled {
			tagger.DryRun()
			fmt.Println("!! Running in a dry run mode")
			fmt.Println("!! No actions will be executed")
		}

		err = tagger.EvaluateRules(resources)
		if err != nil {
			return errors.Wrap(err, "can't eval rules")
		}

		fmt.Println("Evaluating conditions")
		for _, i := range tagger.Matched {
			r := i.Resource
			fmt.Printf("Conditions of [%d] rule(s) matched for [%s] in [%s] with ID %s\n", len(i.TagRules), *r.Name, *r.ResourceGroup, r.ID)
		}

		if len(tagger.Matched) > 0 {
			fmt.Println("\nExecuting actions on matched resources")
			backupFile := azure.NewBackupFromMatched(tagger.Matched, "")
			fmt.Printf("Backup will be saved in: %s\n", backupFile)

			err, ael := tagger.ExecuteActions()

			if err != nil {
				return errors.Wrap(err, "can't exec actions")
			}
			fmt.Println("Executing actions")
			for _, ae := range ael {
				fmt.Printf("Rule [%s] on [%s]\n", ae.RuleName, ae.ResourceID)
				for _, action := range ae.Actions {
					fmt.Printf("Action: [%s] [%s = %s]\n", action.GetType(), action["tag"], action["value"])
				}
			}

		} else {
			fmt.Println("No resources matched your conditions 😫")
		}

		return nil
	},
}
