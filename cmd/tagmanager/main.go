package main

import (
	"fmt"
	"os"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"bitbucket.org/nordcloud/tagmanager/internal/rules"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	usageVerbosity   = "Use for verbose (diagnostic) output"
	usageMappingFile = "Location of the tag rules definition (json)"
	usageDryRun      = "The tagger will not execute any actions"
	usageCommand     = "A mode of operation - choose (rew or check) "
	rewriteCommand   = "rew"
	checkCommand     = "check"
)

var mappingFile string
var dryRunEnabled bool
var verboseEnabled bool
var command string

func init() {

	flag.BoolVar(&verboseEnabled, "verbose", false, usageVerbosity)
	flag.StringVarP(&mappingFile, "map", "m", "", usageMappingFile)
	flag.BoolVar(&dryRunEnabled, "dry", false, usageDryRun)
	flag.StringVarP(&command, "command", "c", checkCommand, usageCommand)
	flag.Parse()

	if verboseEnabled {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	if command == rewriteCommand && mappingFile == "" {
		fmt.Printf("Command rewrite needs a file given (-m)")
		os.Exit(1)
	}

}

func main() {

	switch command {
	case rewriteCommand:
		t, err := rules.NewRulesFromFile(mappingFile)

		if err != nil {
			fmt.Printf("Can't parse rules from %s: %s\n", mappingFile, err)
			os.Exit(1)
		}

		tagger := azure.NewAzureTagger(t)
		tagger.Session, err = session.NewSessionFromFile()
		tagger.DryRun = dryRunEnabled

		if err != nil {
			fmt.Println("Can't create tagger", err)
			os.Exit(1)
		}

		if dryRunEnabled {
			fmt.Println("🤡 🤡  Running in a dry run")
		}

		scanner := azure.NewResourceGroupScanner(tagger.Session)
		res, err := scanner.GetResources()

		if err != nil {
			fmt.Println("can't scan resources", err)
		}

		err = tagger.EvaluateRules(&res)

		if err != nil {
			log.WithError(err).Fatal("can't eval rules")
		}

		for _, i := range tagger.Found {
			r := i.Resource
			fmt.Printf("👍  Rule '%s' found matching resource (%s) with ID = %s\n", i.TagRule.Name, *r.Name, r.ID)
		}
		if len(tagger.Found) > 0 {
			fmt.Println("🔫  Starting executing actions on matched resources")
			err = tagger.ExecuteActions()
			if err != nil {
				fmt.Println("can't exec actions")
				return
			}
		} else {
			fmt.Println("😫  No resources matched your conditions")
		}
	case checkCommand:
		if dryRunEnabled {
			fmt.Println("🤡 🤡  Running in a dry run")
		}
		var err error
		checker := azure.NewAzureChecker()
		checker.Session, err = session.NewSessionFromFile()
		if err != nil {
			fmt.Println("Can't create checker", err)
			os.Exit(1)
		}
		scanner := azure.NewResourceGroupScanner(checker.Session)

		rg := "darek"
		fmt.Println("checking group", rg)
		res, err := scanner.GetResourcesByResourceGroup(rg)
		if err != nil {
			fmt.Println(err)
		}

		nonc := checker.CheckResourceGroup(res)
		for tag, nonrList := range nonc {
			fmt.Printf("Tag with different values: %s\n", tag)
			for _, nonr := range nonrList {
				fmt.Println(nonr)
			}
		}
		// }

	}
}
