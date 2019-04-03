package commands

import (
	"fmt"

	"github.com/pkg/errors"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
)

func Check(c Config) error {
	sess, err := session.NewFromFile()
	if err != nil {
		return errors.Wrap(err, "could not create session")
	}

	checker := azure.NewAzureChecker(sess)
	if c.DryRun {
		checker.DryRun()
		fmt.Println("    Running in a dry run")
	}

	scanner := azure.NewResourceGroupScanner(sess)

	fmt.Println("checking group", c.ResourceGroup)
	res, err := scanner.GetResourcesByResourceGroup(c.ResourceGroup)
	if err != nil {
		return errors.Wrap(err, "could not get resources by group")
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
