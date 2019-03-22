package commands

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func Rewrite(c Config) error {
	t, err := rules.NewFromFile(c.MappingFile)
	if err != nil {
		log.Errorf("Can't parse rules from %s: %s\n", c.MappingFile, err)
	}

	sess, err := session.NewFromFile()
	if err != nil {
		log.Error(errors.Wrap(err, "could not create session"))
	}

	tagger := azure.NewTagger(t, sess)
	if c.DryRun {
		tagger.DryRun()
		fmt.Println("    Running in a dry run")
	}

	if err != nil {
		log.Error(errors.Wrap(err, "Can't create tagger"))
	}

	scanner := azure.NewResourceGroupScanner(tagger.Session)
	res, err := scanner.GetResources()
	if err != nil {
		log.Error(errors.Wrap(err, "can't scan resources"))
	}

	err = tagger.EvaluateRules(res)
	if err != nil {
		log.Error(errors.Wrap(err, "can't eval rules"))
	}

	for _, i := range tagger.Found {
		r := i.Resource
		fmt.Printf("	👍Rule '%s' found matching resource (%s) with ID = %s\n", i.TagRule.Name, *r.Name, r.ID)
	}

	if len(tagger.Found) > 0 {
		fmt.Println("🔫  Starting executing actions on matched resources")
		if err := tagger.ExecuteActions(); err != nil {
			log.Error(errors.Wrap(err, "can't exec actions"))
		}
	} else {
		fmt.Println("😫  No resources matched your conditions")
	}

	return nil
}