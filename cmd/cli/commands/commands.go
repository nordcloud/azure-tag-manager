package commands

import "github.com/pkg/errors"

type Pool struct {
	Commands map[string]Command
}

func (p *Pool) Execute(cfg Config, cmdName string) error {
	cmd, ok := p.Commands[cmdName]
	if !ok {
		return errors.New("provided command does not exist")
	}

	if err := cmd.Validate(cfg); err != nil {
		return errors.Wrap(err, "config is not valid")
	}

	return cmd.Execute(cfg)
}

type Command interface {
	Execute(Config) error
	Validate(Config) error
}
