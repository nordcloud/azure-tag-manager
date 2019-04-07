package main

import (
	"bitbucket.org/nordcloud/tagmanager/cmd/cli/commands"
	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	usageVerbosity     = "Use for verbose (diagnostic) output"
	usageMappingFile   = "Location of the tag rules definition (json)"
	usageDryRun        = "The tagger will not execute any actions"
	usageCommand       = "A mode of operation - choose (rew or check)"
	usageResourceGroup = "Specifies resource group"
	usageRestoreFile   = "Specify the location of the restore file"
)

const (
	commandRewrite = "rew"
	commandCheck   = "check"
	commandRestore = "restore"
)

var (
	mappingFile    string
	dryRunEnabled  bool
	verboseEnabled bool
	command        string
	resourceGroup  string
	restoreFile    string
)

func init() {
	flag.BoolVar(&verboseEnabled, "verbose", false, usageVerbosity)
	flag.StringVarP(&mappingFile, "map", "m", "", usageMappingFile)
	flag.BoolVar(&dryRunEnabled, "dry", false, usageDryRun)
	flag.StringVarP(&command, "command", "c", commandCheck, usageCommand)
	flag.StringVarP(&resourceGroup, "resourceGroup", "r", "", usageResourceGroup)
	flag.StringVarP(&restoreFile, "restoreFile", "f", "", usageRestoreFile)

	flag.Parse()

	if verboseEnabled {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
}

func main() {
	cfg := commands.Config{
		MappingFile:   mappingFile,
		DryRun:        dryRunEnabled,
		ResourceGroup: resourceGroup,
		RestoreFile:   restoreFile,
	}

	pool := commands.Pool{
		Commands: map[string]commands.Command{
			commandCheck:   &commands.CheckCommand{},
			commandRewrite: &commands.RewriteCommand{},
			commandRestore: &commands.RestoreCommand{},
		},
	}

	if err := pool.Execute(cfg, command); err != nil {
		log.Fatal(errors.Wrap(err, "could not execute command"))
	}
}
