package rules

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

func NewRulesFromFile(filename string) (*TagRules, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "error opening the file")
	}
	return NewRulesFromString(string(dat))
}

func NewRulesFromString(rulesDef string) (*TagRules, error) {
	return parseRulesDefinitions(rulesDef)
}

func parseRulesDefinitions(rules string) (*TagRules, error) {
	var rulesDef TagRules
	byt := []byte(rules)
	if err := json.Unmarshal(byt, &rulesDef); err != nil {
		return nil, errors.Wrap(err, "can't unmarshal rules")
	}

	return &rulesDef, nil
}

type TagRules struct {
	DryRun *bool  `json:"dryrun,omitempty"`
	Rules  []Rule `json:"rules"`
}

type Rule struct {
	Conditions []ConditionItem `json:"conditions"`
	Actions    []ActionItem    `json:"actions"`
}

type ConditionItem map[string]string

type ActionItem map[string]string

func (p ConditionItem) GetType() string {
	if val, ok := p["type"]; ok {
		return val
	}
	return ""
}

func (p ActionItem) GetType() string {
	if val, ok := p["type"]; ok {
		return val
	}
	return ""
}
