package commands

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/pkg/errors"
)

type RewriteCommand struct{}

func (r *RewriteCommand) Execute(cfg Config) error {
	t, err := rules.NewFromFile(cfg.MappingFile)
	if err != nil {
		return errors.Wrapf(err, "Can't parse rules from %s", cfg.MappingFile)
	}

	sess, err := session.NewFromFile()
	if err != nil {
		return errors.Wrap(err, "could not create session")
	}

	tagger := azure.NewTagger(t, sess)
	if cfg.DryRun {
		tagger.DryRun()
		fmt.Println("    Running in a dry run")
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

	for _, i := range tagger.Found {
		r := i.Resource
		fmt.Printf("	👍Rule '%s' found matching resource (%s) with ID = %s\n", i.TagRule.Name, *r.Name, r.ID)
	}

	if len(tagger.Found) > 0 {
		fmt.Println("🔫  Starting executing actions on matched resources")
		if err := tagger.ExecuteActions(); err != nil {
			return errors.Wrap(err, "can't exec actions")
		}
	} else {
		fmt.Println("😫  No resources matched your conditions")
	}

	return nil
}

func (r *RewriteCommand) Validate(cfg Config) error {
	if cfg.MappingFile == "" {
		return errors.New("need a mapping file given (-m)")
	}

	return nil
}
