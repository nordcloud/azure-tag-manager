package commands

import (
	"fmt"

	"github.com/pkg/errors"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
)

type CheckCommand struct{}

func (c *CheckCommand) Execute(cfg Config) error {
	sess, err := session.NewFromFile()
	if err != nil {
		return errors.Wrap(err, "could not create session")
	}

	fmt.Println("checking group", cfg.ResourceGroup)
	scanner := azure.NewResourceGroupScanner(sess)
	res, err := scanner.GetResourcesByResourceGroup(cfg.ResourceGroup)
	if err != nil {
		return errors.Wrap(err, "could not get resources by group")
	}

	checker := azure.NewAzureChecker(sess)
	if cfg.DryRun {
		checker.DryRun()
		fmt.Println("    Running in a dry run")
	}

	nonc := checker.CheckResourceGroup(res)
	for tag, nonrList := range nonc {
		fmt.Printf("Tag with different values: %s\n", tag)
		for _, nonr := range nonrList {
			fmt.Println(nonr)
		}
	}

	return nil
}

func (c *CheckCommand) Validate(cfg Config) error {
	if cfg.ResourceGroup == "" {
		return errors.New("need resource group to be specified (-r)")
	}

	return nil
}
