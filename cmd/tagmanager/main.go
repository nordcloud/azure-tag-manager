package main

import (
	"github.com/pkg/errors"

	"bitbucket.org/nordcloud/tagmanager/cmd/tagmanager/commands"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	usageVerbosity   = "Use for verbose (diagnostic) output"
	usageMappingFile = "Location of the tag rules definition (json)"
	usageDryRun      = "The tagger will not execute any actions"
	usageCommand     = "A mode of operation - choose (rew or check)"

	rewriteCommand = "rew"

	checkCommand = "check"
)

var (
	mappingFile    string
	dryRunEnabled  bool
	verboseEnabled bool
	command        string
)

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
		log.Error("Command rewrite needs a file given (-m)")
	}
}

func main() {
	c := commands.Config{
		MappingFile: mappingFile,
		DryRun:      dryRunEnabled,
	}

	switch command {
	case rewriteCommand:
		err := commands.Rewrite(c)
		if err != nil {
			log.Error(errors.Wrap(err, "could not execute rewrite command"))
		}
	case checkCommand:
		err := commands.Check(c)
		if err != nil {
			log.Error(errors.Wrap(err, "could not execute check command"))
		}
	}
}