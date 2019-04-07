package azure

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func NewTagger(ruleDef rules.TagRules, session *session.AzureSession) *Tagger {
	tagger := Tagger{
		Session: session,
		Rules:   ruleDef,
		Found:   make(map[string]Found),
	}

	tagger.InitActionMap()
	tagger.InitCondMap()

	return &tagger
}

// Found stores
type Found struct {
	Actions  []rules.ActionItem
	Resource Resource
	TagRule  rules.Rule
}

type Tagger struct {
	Session   *session.AzureSession
	Found     map[string]Found
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

func (t Tagger) ExecuteActions() error {
	for resID, found := range t.Found {
		fmt.Printf("Executing actions of rule [%s] on [%s]\n", found.TagRule.Name, resID)
		for _, action := range found.Actions {
			if t.dryRun == true {
				fmt.Printf("(dryRun) [%s] Action [%s] (%s=%s)\n", found.TagRule.Name, action.GetType(), action["tag"], action["value"])
			} else {
				fmt.Printf("!!! [%s] Action [%s] (%s=%s)\n", found.TagRule.Name, action.GetType(), action["tag"], action["value"])
				resource := Resource{ID: resID}
				err := t.Execute(&resource, action)
				if err != nil {
					log.Errorf("Can't execute action [%s] on [%s]\n", action.GetType(), resource.ID)
				}
			}
		}
	}

	return nil
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
				found := Found{Actions: y.Actions, Resource: resource, TagRule: y}
				t.Found[resource.ID] = found
			}
		}
	}

	return nil
}
