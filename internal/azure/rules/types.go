package rules

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

func NewFromFile(filename string) (TagRules, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return TagRules{}, errors.Wrap(err, "error opening the file")
	}

	return NewFromString(string(dat))
}

func NewFromString(rulesDef string) (TagRules, error) {
	return parseRulesDefinitions(rulesDef)
}

type TagRules struct {
	DryRun *bool  `json:"dryrun,omitempty"`
	Rules  []Rule `json:"rules"`
}

type Rule struct {
	Name       string          `json:"name,omitempty"`
	Conditions []ConditionItem `json:"conditions"`
	Actions    []ActionItem    `json:"actions"`
}

type ConditionItem map[string]string

func (p ConditionItem) GetType() string {
	if val, ok := p["type"]; ok {
		return val
	}
	return ""
}

type ActionItem map[string]string

func (p ActionItem) GetType() string {
	if val, ok := p["type"]; ok {
		return val
	}
	return ""
}

func parseRulesDefinitions(rules string) (TagRules, error) {
	var rulesDef TagRules
	byt := []byte(rules)
	if err := json.Unmarshal(byt, &rulesDef); err != nil {
		return TagRules{}, errors.Wrap(err, "can't unmarshal rules")
	}

	return rulesDef, nil
}
