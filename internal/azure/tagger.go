package azure

import (
	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//ActionExecution stores information about execution of actions of a rule
type ActionExecution struct {
	ResourceID string
	RuleName   string
	Actions    []rules.ActionItem
}

//NewTagger creates tagger
func NewTagger(ruleDef rules.TagRules, session *session.AzureSession) *Tagger {
	tagger := Tagger{
		Session: session,
		Rules:   ruleDef,
		Matched: make(map[string]Matched),
	}
	tagger.InitActionMap()
	tagger.InitCondMap()

	return &tagger
}

// Matched stores
type Matched struct {
	Resource Resource
	TagRules []rules.Rule
}

type Tagger struct {
	Session   *session.AzureSession
	Matched   map[string]Matched
	Rules     rules.TagRules
	condMap   condFuncMap
	actionMap actionFuncMap
	dryRun    bool
}

func (t *Tagger) DryRun() {
	t.dryRun = true
}

func (t *Tagger) InitActionMap() {
	t.actionMap = actionFuncMap{}
	t.actionMap["addTag"] = func(p map[string]string, data *Resource) error {
		err := t.createOrUpdateTag(data.ID, p["tag"], p["value"])
		if err != nil {
			return errors.Wrapf(err, "Action addTag did not succeed for resource %s", data.ID)
		}

		return nil
	}

}

func (t *Tagger) InitCondMap() {
	t.condMap = condFuncMap{}
	t.condMap["noTags"] = func(p map[string]string, data *Resource) bool {
		if len(data.Tags) == 0 {
			return true
		}
		return false
	}

	t.condMap["tagEqual"] = func(p map[string]string, data *Resource) bool {
		tags := data.Tags
		if len(tags) == 0 {
			return false
		}
		for k, tag := range tags {
			if p["tag"] == k && p["value"] == *tag {
				return true
			}
		}
		return false
	}

	t.condMap["tagNotEqual"] = func(p map[string]string, data *Resource) bool {
		tags := data.Tags
		if len(tags) == 0 {
			return false
		}
		for k, tag := range tags {
			if p["tag"] == k && p["value"] != *tag {
				return true
			}
		}
		return false
	}

	t.condMap["tagExists"] = func(p map[string]string, data *Resource) bool {
		tags := data.Tags
		if len(tags) == 0 {
			return false
		}
		if _, ok := tags[p["tag"]]; ok {
			return true
		}
		return false

	}

	t.condMap["tagNotExists"] = func(p map[string]string, data *Resource) bool {
		tags := data.Tags
		if len(tags) == 0 {
			return true
		}
		if _, ok := tags[p["tag"]]; !ok {
			return true
		}
		return false
	}

	t.condMap["regionEqual"] = func(p map[string]string, data *Resource) bool {
		if p["region"] == data.Region {
			return true
		}
		return false
	}

	t.condMap["regionNotEqual"] = func(p map[string]string, data *Resource) bool {
		if p["region"] != data.Region {
			return true
		}
		return false
	}

	t.condMap["rgEqual"] = func(p map[string]string, data *Resource) bool {
		if p["resourceGroup"] == *data.ResourceGroup {
			return true
		}
		return false
	}

	t.condMap["rgNotEqual"] = func(p map[string]string, data *Resource) bool {
		if p["resourceGroup"] != *data.ResourceGroup {
			return true
		}
		return false
	}
}

func (t Tagger) ExecuteActions() (error, []ActionExecution) {
	ael := make([]ActionExecution, 0)
	for resID, matched := range t.Matched {

		for _, rule := range matched.TagRules {
			ae := ActionExecution{
				ResourceID: resID,
				RuleName:   rule.Name,
				Actions:    rule.Actions,
			}
			for _, action := range rule.Actions {
				if t.dryRun == true {
				} else {
					resource := Resource{ID: resID}
					err := t.Execute(&resource, action)
					if err != nil {
						log.Errorf("Can't execute action [%s] on [%s]\n", action.GetType(), resource.ID)
					}
				}
			}
			ael = append(ael, ae)
		}
	}
	return nil, ael
}

// EvaluateRules iterates over all rules and resources and checks which conditions are true.
func (t Tagger) EvaluateRules(resources []Resource) error {
	var evaled bool
	for _, resource := range resources {
		evaled = true
		log.Debugf("🔍  Checking resource: %s (%s) \n", *resource.Name, resource.ID)
		for _, y := range t.Rules.Rules {
			for _, cond := range y.Conditions {
				evaled = t.Eval(&resource, cond)
				if !evaled {
					break
				}
			}

			if evaled {
				if val, ok := t.Matched[resource.ID]; ok {
					matched := Matched{Resource: resource, TagRules: append(val.TagRules, y)}
					t.Matched[resource.ID] = matched
				} else {
					matched := Matched{Resource: resource, TagRules: []rules.Rule{y}}
					t.Matched[resource.ID] = matched
				}
			}
		}
	}

	return nil
}
