package azure

import (
	"fmt"

	"bitbucket.org/nordcloud/pantageusz/internal/azure/session"
	"bitbucket.org/nordcloud/pantageusz/internal/rules"
	tag "bitbucket.org/nordcloud/pantageusz/internal/tagger"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Found struct {
	Actions  []rules.ActionItem
	Resource tag.TaggableResource
}

type Tagger struct {
	Session   *session.AzureSession
	Found     map[string]Found
	condMap   tag.CondFuncMap
	actionMap tag.ActionFuncMap
	dryRun    bool
	Rules     rules.TagRules
}

func NewAzureTagger(ruleDef *rules.TagRules) (*Tagger, error) {
	var err error
	tagger := &Tagger{}
	tagger.Session, err = session.NewSessionFromFile()

	if err != nil {
		return nil, errors.Wrap(err, "can't create tagger")
	}
	tagger.Rules = *ruleDef
	tagger.dryRun = *ruleDef.DryRun
	tagger.InitActionMap()
	tagger.InitCondMap()
	tagger.Found = map[string]Found{}
	return tagger, nil
}

func (t *Tagger) InitActionMap() {
	t.actionMap = tag.ActionFuncMap{}
	t.actionMap["addTag"] = func(p map[string]string, data *tag.TaggableResource) error {
		err := t.createOrUpdateTag(data.ID, p["tag"], p["value"])
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Action addTag did not succeed for resource %s", data.ID))
		}
		return err
	}
}

func (t *Tagger) InitCondMap() {
	t.condMap = tag.CondFuncMap{}
	t.condMap["tagValue"] = func(p map[string]string, data *tag.TaggableResource) bool {
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

	t.condMap["tagExists"] = func(p map[string]string, data *tag.TaggableResource) bool {
		tags := data.Tags
		if len(tags) == 0 {
			return false
		}

		if _, ok := tags[p["tag"]]; ok {
			return true
		}
		return false
	}

	t.condMap["tagNotExists"] = func(p map[string]string, data *tag.TaggableResource) bool {
		tags := data.Tags
		if len(tags) == 0 {
			return true
		}
		if _, ok := tags[p["tag"]]; !ok {
			return true
		}
		return false
	}

	t.condMap["regionEqual"] = func(p map[string]string, data *tag.TaggableResource) bool {
		if p["region"] == data.Region {
			return true
		}
		return false
	}

	t.condMap["regionNotEqual"] = func(p map[string]string, data *tag.TaggableResource) bool {
		if p["region"] != data.Region {
			return true
		}
		return false
	}

}

func (t Tagger) ExecuteActions() error {
	for resID, found := range t.Found {
		for _, action := range found.Actions {
			if t.dryRun == true {
				log.Printf("\t🔥  DryRun Firing action %s on resource %s\n", action.GetType(), resID)
			} else {
				log.Printf("\t🔥  Firing action %s on resource %s\n", action.GetType(), resID)
				resource := tag.TaggableResource{ID: resID}
				err := t.Execute(&resource, action)
				if err != nil {
					fmt.Printf("Can't fire rule %s on %s\n", action.GetType(), resource.ID)
				}
			}
		}
	}
	return nil
}
func (t Tagger) EvaluteRules(resources *[]tag.TaggableResource) error {
	var evaled bool
	// iterate over resources

	for _, resource := range *resources {
		evaled = true
		// log.Infof("🔍 checking resource: %s\n", resource.ID)
		for _, y := range t.Rules.Rules {
			for _, cond := range y.Conditions {

				evaled = t.Eval(&resource, cond)

				if !evaled {
					break
				}
			}
			if evaled {
				log.Infof("👍  Rules %t for %s\n", evaled, *resource.Name)
				found := Found{Actions: y.Actions, Resource: resource}
				t.Found[resource.ID] = found
			}
		}
	}
	return nil
}
