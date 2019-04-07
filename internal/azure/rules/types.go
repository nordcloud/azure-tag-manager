package rules

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"unicode"

	"github.com/ghodss/yaml"

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

var jsonPrefix = []byte("{")

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

func ToJSON(data []byte) ([]byte, error) {
	if hasJSONPrefix(data) {
		return data, nil
	}
	return yaml.YAMLToJSON(data)
}
