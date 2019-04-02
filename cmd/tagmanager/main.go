package main

import (
	"github.com/pkg/errors"

	"bitbucket.org/nordcloud/tagmanager/cmd/tagmanager/commands"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	usageVerbosity     = "Use for verbose (diagnostic) output"
	usageMappingFile   = "Location of the tag rules definition (json)"
	usageDryRun        = "The tagger will not execute any actions"
	usageCommand       = "A mode of operation - choose (rew or check)"
	usageResourceGroup = "Specifies resource group"
)

const (
	commandRewrite = "rew"
	commandCheck   = "check"
)

var (
	mappingFile    string
	dryRunEnabled  bool
	verboseEnabled bool
	command        string
	resourceGroup  string
)

func init() {
	flag.BoolVar(&verboseEnabled, "verbose", false, usageVerbosity)
	flag.StringVarP(&mappingFile, "map", "m", "", usageMappingFile)
	flag.BoolVar(&dryRunEnabled, "dry", false, usageDryRun)
	flag.StringVarP(&command, "command", "c", commandCheck, usageCommand)
	flag.StringVarP(&resourceGroup, "resourceGroup", "r", "", usageResourceGroup)
	flag.Parse()

	if verboseEnabled {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

	if command == commandRewrite && mappingFile == "" {
		log.Fatal("Command rewrite needs a file given (-m)")
	}

	if command == commandCheck && resourceGroup == "" {
		log.Fatal("Command check needs resource group to be specified (-r)")
	}
}

func main() {
	c := commands.Config{
		MappingFile:   mappingFile,
		DryRun:        dryRunEnabled,
		ResourceGroup: resourceGroup,
	}

	switch command {
	case commandRewrite:
		err := commands.Rewrite(c)
		if err != nil {
			log.Fatal(errors.Wrap(err, "could not execute rewrite command"))
		}
	case commandCheck:
		err := commands.Check(c)
		if err != nil {
			log.Fatal(errors.Wrap(err, "could not execute check command"))
		}
	}
}
