package rules

import (
	"reflect"
	"testing"
)

const (
	yamlTwo = `
---
rules:
- name: name
  conditions:
  - type: tagEqual
    tag: test
    value: test
  - type: tagExists
    tag: test
  actions:
  - type: addTag
    tag: test
    value: test
`
	one = `{
		"dryrun": true,
		  "rules":  [
			{
				"name": "Tag me this",
				"conditions": [
					{"type": "tagEqual", "tag": "test", "value" : "test"},
					{"type": "tagExists", "tag": "test"},
				], 
				"actions": [
					{"type": "addTag", "tag": "test", "value": "test" },
				]
			},
			{
				"name": "Tag me this2",
				"conditions": [
					{"type": "tagEqual", "tag": "test", "value" : "test"},
					{"type": "tagExists", "tag": "test"},
				], 
				"actions": [
					{"type": "addTag", "tag": "test", "value": "test" },
				]
			}
			]
		}`
	two = `{ "rules":  [
				{
					"name": "name",
					"conditions": [
						{"type": "tagEqual", "tag": "test", "value" : "test"},
						{"type": "tagExists", "tag": "test"}
					], 
					"actions": [
						{"type": "addTag", "tag": "test", "value": "test" }
					]
				}
				]
			}`
	empty      = `{}`
	onlyDryRun = `{"dryrun": true}`
	wrongJson  = `{ew2`
	wrongYaml  = `223322`
)

var (
	twoRulesWant = TagRules{Rules: []Rule{
		Rule{Name: "name", Conditions: []ConditionItem{
			ConditionItem{"type": "tagEqual", "tag": "test", "value": "test"},
			ConditionItem{"type": "tagExists", "tag": "test"},
		},
			Actions: []ActionItem{
				ActionItem{"type": "addTag", "tag": "test", "value": "test"},
			},
		},
	}}
)

var dryRunFalse bool = false
var dryRunTrue bool = true

func TestNewFromString(t *testing.T) {
	type args struct {
		rulesDef string
	}
	tests := []struct {
		name    string
		args    args
		want    TagRules
		wantErr bool
	}{
		{name: "empty", args: args{rulesDef: empty}, want: TagRules{}, wantErr: false},
		{name: "only dryrun defined", args: args{rulesDef: onlyDryRun}, want: TagRules{DryRun: &dryRunTrue}, wantErr: false},
		{name: "one rule", args: args{rulesDef: two}, want: twoRulesWant, wantErr: false},
		{name: "one rule yaml", args: args{rulesDef: yamlTwo}, want: twoRulesWant, wantErr: false},
		{name: "wrong json", args: args{rulesDef: wrongJson}, want: TagRules{}, wantErr: true},
		{name: "wrong yaml", args: args{rulesDef: wrongYaml}, want: TagRules{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewFromString(tt.args.rulesDef)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
