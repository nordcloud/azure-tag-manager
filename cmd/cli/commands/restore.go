package commands

import (
	"fmt"

	"github.com/pkg/errors"

	"bitbucket.org/nordcloud/tagmanager/internal/azure"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
)

type RestoreCommand struct{}

func (c *RestoreCommand) Execute(cfg Config) error {

	sess, err := session.NewFromFile()
	if err != nil {
		return errors.Wrap(err, "could not create session")
	}

	fmt.Printf("Restoring tags from: [%s]\n", cfg.RestoreFile)

	restorer := azure.NewRestorerFromFile(cfg.RestoreFile, sess, false)
	err = restorer.Restore()

	if err != nil {
		return errors.Wrap(err, "could not restore backup")
	}
	return nil
}

func (c *RestoreCommand) Validate(cfg Config) error {
	if cfg.RestoreFile == "" {
		return errors.New("Restorefile needs to be specified (-f)")
	}

	return nil
}
