package commands

import (
	"fmt"

	"github.com/pkg/errors"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
)

func Check(c Config) error {
	if c.DryRun {
		fmt.Println("🤡 🤡  Running in a dry run")
	}

	sess, err := session.NewFromFile()
	if err != nil {
		return errors.Wrap(err, "could not create session")
	}

	checker := azure.NewAzureChecker(sess)
	scanner := azure.NewResourceGroupScanner(sess)

	// @TODO
	rg := "darek"
	fmt.Println("checking group", rg)
	res, err := scanner.GetResourcesByResourceGroup(rg)
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
