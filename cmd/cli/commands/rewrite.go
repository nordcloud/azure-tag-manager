package commands

import (
	"fmt"

	"github.com/nordcloud/azure-tag-manager/internal/azure"
	"github.com/nordcloud/azure-tag-manager/internal/azure/rules"
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ActionExecution struct {
}

const (
	usageMappingFile = "Location of the tag rules definition (json)"
	usageDryRun      = "The tagger will not execute any actions"
)

var (
	mappingFile   string
	dryRunEnabled bool
)

func init() {
	rootCmd.AddCommand(rewriteCommand)
	rewriteCommand.Flags().StringVarP(&mappingFile, "map", "m", "", usageMappingFile)
	rewriteCommand.MarkFlagRequired("map")
	rewriteCommand.Flags().BoolVar(&dryRunEnabled, "dry", false, usageDryRun)
}

var rewriteCommand = &cobra.Command{
	Use:   "rewrite",
	Short: "Rewrite tags based on rules from a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t, err := rules.NewFromFile(mappingFile)
		if err != nil {
			return errors.Wrapf(err, "Can't parse rules from %s", mappingFile)
		}

		sess, err := session.NewFromFile()
		if err != nil {
			return errors.Wrap(err, "Could not create session")
		}

		tagger := azure.NewTagger(t, sess)
		if dryRunEnabled {
			tagger.DryRun()
			fmt.Println("!! Running in a dry run mode")
			fmt.Println("!! No actions will be executed")
		}

		if err != nil {
			return errors.Wrap(err, "Can't create tagger")
		}

		scanner := azure.NewResourceGroupScanner(tagger.Session)
		res, err := scanner.GetResources()
		if err != nil {
			return errors.Wrap(err, "can't scan resources")
		}

		tagger.EvaluateRules(res)

		fmt.Println("Evaluating conditions")
		for _, i := range tagger.Matched {
			r := i.Resource
			fmt.Printf("Conditions of [%d] rule(s) matched for [%s] in [%s] with ID %s\n", len(i.TagRules), *r.Name, *r.ResourceGroup, r.ID)
		}

		if len(tagger.Matched) > 0 {
			fmt.Println("\nExecuting actions on matched resources")
			backupFile := azure.NewBackupFromMatched(tagger.Matched, "")
			fmt.Printf("Backup will be saved in: %s\n", backupFile)

			ael, err := tagger.ExecuteActions()

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
