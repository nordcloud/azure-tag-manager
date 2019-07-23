package rules

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"unicode"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
)

// NewFromFile reads filename and returns TagRules
func NewFromFile(filename string) (TagRules, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return TagRules{}, errors.Wrap(err, "error opening the file")
	}

	return NewFromString(string(dat))
}

// NewFromString parses rulesDef and returns TagRules
func NewFromString(rulesDef string) (TagRules, error) {
	return parseRulesDefinitions(rulesDef)
}

// TagRules represents rules parsed from a rules definition
type TagRules struct {
	DryRun *bool  `json:"dryrun,omitempty"`
	Rules  []Rule `json:"rules"`
}

// Rule represnts single rule
type Rule struct {
	Name       string          `json:"name,omitempty"`
	Conditions []ConditionItem `json:"conditions"`
	Actions    []ActionItem    `json:"actions"`
}

// ConditionItem represnts one condition
type ConditionItem map[string]string

// GetType retrurn the type of the condition
func (p ConditionItem) GetType() string {
	if val, ok := p["type"]; ok {
		return val
	}
	return ""
}

// ActionItem represnts a single action
type ActionItem map[string]string

// GetType retrurn the type of the action
func (p ActionItem) GetType() string {
	if val, ok := p["type"]; ok {
		return val
	}
	return ""
}

var jsonPrefix = []byte("{")

func parseRulesDefinitions(rules string) (TagRules, error) {
	var rulesDef TagRules
	byt := []byte(rules)
	if hasJSONPrefix(byt) {
		if err := json.Unmarshal(byt, &rulesDef); err != nil {
			return TagRules{}, errors.Wrap(err, "can't unmarshal json rules")
		}
	} else {
		if err := yaml.Unmarshal(byt, &rulesDef); err != nil {
			return TagRules{}, errors.Wrap(err, "can't unmarshal yaml rules")
		}
	}
	return rulesDef, nil
}

// hasJSONPrefix returns true if the provided buffer appears to start with
// a JSON open brace.
func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}
