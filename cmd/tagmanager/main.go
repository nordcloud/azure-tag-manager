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
)

var mappingFile string
var dryRunEnabled bool
var verboseEnabled bool

func init() {

	flag.BoolVar(&verboseEnabled, "verbose", false, usageVerbosity)
	flag.StringVarP(&mappingFile, "map", "m", "", usageMappingFile)
	flag.BoolVar(&dryRunEnabled, "dry", false, usageDryRun)
	flag.Parse()

	if verboseEnabled {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

}

func main() {

	if mappingFile == "" {
		// if len(os.Args) < 2 {
		fmt.Println("Mapping file not given")
		os.Exit(1)
	}

	t, err := rules.NewRulesFromFile(mappingFile)

	if err != nil {
		fmt.Printf("Can't parse rules from %s: %s\n", mappingFile, err)
		os.Exit(1)
	}

	tagger, err := azure.NewAzureTagger(t)
	tagger.Session, err = session.NewSessionFromFile()
	tagger.DryRun = dryRunEnabled

	if err != nil {
		log.WithError(err).Fatal("Can't create tagger")
	}

	scanner := azure.ResourceGroupScanner{Session: tagger.Session}
	res, err := scanner.GetResources()

	if err != nil {
		log.WithError(err).Fatalf("can't scan resources")
	}

	err = tagger.EvaluateRules(&res)
	if err != nil {
		log.WithError(err).Fatal("can't eval rules")
	}

	for _, i := range tagger.Found {
		r := i.Resource
		log.Printf("👍  Rule '%s' found matching resource (%s) with ID = %s\n", i.TagRule.Name, *r.Name, r.ID)
	}
	if len(tagger.Found) > 0 {
		log.Println("🔫  Starting executing actions on matched resources")
		err = tagger.ExecuteActions()
	} else {
		log.Println("😫  No resources matched your conditions")
	}

	if err != nil {
		log.WithError(err).Fatal("can't exec actions")
	}
}
