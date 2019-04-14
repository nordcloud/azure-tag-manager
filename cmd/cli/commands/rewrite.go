package commands

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/pkg/errors"
)

type ActionExecution struct {
}

type RewriteCommand struct{}

func (r *RewriteCommand) Execute(cfg Config) error {
	t, err := rules.NewFromFile(cfg.MappingFile)
	if err != nil {
		return errors.Wrapf(err, "Can't parse rules from %s", cfg.MappingFile)
	}

	sess, err := session.NewFromFile()
	if err != nil {
		return errors.Wrap(err, "Could not create session")
	}

	tagger := azure.NewTagger(t, sess)
	if cfg.DryRun {
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

	err = tagger.EvaluateRules(res)
	if err != nil {
		return errors.Wrap(err, "can't eval rules")
	}

	fmt.Println("Evaluating conditions")
	for _, i := range tagger.Matched {
		r := i.Resource
		fmt.Printf("Conditions of rule [%s] matched [%s] in [%s] with ID %s\n", i.TagRule.Name, *r.Name, *r.ResourceGroup, r.ID)
	}

	if len(tagger.Matched) > 0 {
		fmt.Println("\nExecuting actions on matched resources")
		backupFile := azure.NewBackupFromMatched(tagger.Matched, "")
		fmt.Printf("Backup will be saved in: %s\n", backupFile)

		err, ael := tagger.ExecuteActions()

		if err != nil {
			return errors.Wrap(err, "can't exec actions")
		}

		for _, ae := range ael {
			fmt.Println("Action executions")
			fmt.Printf("Rule [%s] on [%s]\n", ae.RuleName, ae.ResourceID)
			for _, action := range ae.Actions {
				fmt.Printf("Action: [%s] [%s = %s]\n", action.GetType(), action["tag"], action["value"])
			}
		}

	} else {
		fmt.Println("No resources matched your conditions 😫")
	}
	return nil
}

func (r *RewriteCommand) Validate(cfg Config) error {
	if cfg.MappingFile == "" {
		return errors.New("need a mapping file given (-m)")
	}

	return nil
}
