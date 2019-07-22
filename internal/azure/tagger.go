package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources/resourcesapi"
	"github.com/nordcloud/azure-tag-manager/internal/azure/rules"
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Tagger reprents the maing tagging element
type Tagger struct {
	Session         *session.AzureSession
	Matched         map[string]Matched
	Rules           rules.TagRules // list of rules
	condMap         condFuncMap    // map of implementation of conditions
	actionMap       actionFuncMap  // map of implementation of actions
	dryRun          bool           // if true, actions will not be executed
	ResourcesClient resourcesapi.ClientAPI
}

// Matched represents rules that mathc for a resource
type Matched struct {
	Resource Resource
	TagRules []rules.Rule
}

//ActionExecution stores information about execution of actions of a rule
type ActionExecution struct {
	ResourceID string
	RuleName   string
	Actions    []rules.ActionItem
}

//NewTagger creates tagger
func NewTagger(ruleDef rules.TagRules, session *session.AzureSession) *Tagger {
	grClient := resources.NewClient(session.SubscriptionID)
	grClient.Authorizer = session.Authorizer

	tagger := Tagger{
		Session:         session,
		Rules:           ruleDef,
		Matched:         make(map[string]Matched),
		ResourcesClient: &grClient,
	}

	tagger.InitActionMap()
	tagger.InitCondMap()

	return &tagger
}

// DryRun returns true if the check should be simulated
func (t *Tagger) DryRun() {
	t.dryRun = true
}

// InitActionMap initializes action map with supported actions
func (t *Tagger) InitActionMap() {
	t.actionMap = actionFuncMap{}
	t.actionMap["addTag"] = func(p map[string]string, data *Resource) error {
		err := t.createOrUpdateTag(data.ID, p["tag"], p["value"])
		if err != nil {
			return errors.Wrapf(err, "Action addTag failed for resource %s", data.ID)
		}

		return nil
	}

	t.actionMap["delTag"] = func(p map[string]string, data *Resource) error {
		err := t.deleteTag(data.ID, p["tag"])
		if err != nil {
			return errors.Wrapf(err, "Action delTag failed for resource %s", data.ID)
		}
		return nil
	}

	t.actionMap["cleanTags"] = func(p map[string]string, data *Resource) error {
		err := t.deleteAllTags(data.ID)
		if err != nil {
			return errors.Wrapf(err, "Action cleanTags failed for resource %s", data.ID)
		}
		return nil
	}

}

// InitCondMap initializes conditions map with supported conditions
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

	t.condMap["resEqual"] = func(p map[string]string, data *Resource) bool {
		if p["resourceGroup"] != *data.ResourceGroup {
			return true
		}
		return false
	}
}

// ExecuteActions executes all actions based on definitions of rules. It resturns list of executed actions
func (t *Tagger) ExecuteActions() ([]ActionExecution, error) {
	ael := make([]ActionExecution, 0)
	for resID, matched := range t.Matched {
		for _, rule := range matched.TagRules {
			ae := ActionExecution{
				ResourceID: resID,
				RuleName:   rule.Name,
				Actions:    rule.Actions,
			}
			for _, action := range rule.Actions {
				if t.dryRun != true {
					resource := Resource{ID: resID}
					err := t.Execute(&resource, action)
					if err != nil {
						msg := fmt.Sprintf("ExecuteActions(): Execute() failed Can't execute action [%s] on [%s], [%s]\n", action.GetType(), resource.ID, err)
						return []ActionExecution{}, errors.New(msg)
					}
				}
			}
			ael = append(ael, ae)
		}
	}
	return ael, nil
}

// EvaluateRules iterates over all rules and resources and checks which conditions are true.
func (t Tagger) EvaluateRules(resources []Resource) {
	var evaled bool

	for _, resource := range resources {
		evaled = true
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
}

func (t Tagger) deleteAllTags(id string) error {
	genericResource := resources.GenericResource{
		Tags: make(map[string]*string),
	}

	_, err := t.ResourcesClient.UpdateByID(context.Background(), id, genericResource)
	if err != nil {
		return errors.Wrapf(err, "deleteAllTags(id=%s): UpdateByID() failed", id)
	}
	return nil
}

func (t Tagger) deleteTag(id, tag string) error {

	r, err := t.ResourcesClient.GetByID(context.Background(), id)
	if err != nil {
		return errors.Wrapf(err, "deleteTag(id=%s, tag=%s): GetByID failed", id, tag)
	}

	if _, ok := r.Tags[tag]; !ok {
		return nil
	}

	delete(r.Tags, tag)
	genericResource := resources.GenericResource{
		Tags: r.Tags,
	}

	_, err = t.ResourcesClient.UpdateByID(context.Background(), id, genericResource)
	if err != nil {
		return errors.Wrapf(err, "deleteTag(id=%s, tag=%s): UpdateByID() failed", id, tag)
	}
	return err
}

func (t Tagger) createOrUpdateTag(id, tag, value string) error {

	r, err := t.ResourcesClient.GetByID(context.Background(), id)
	if err != nil {
		return errors.Wrap(err, "cannot get resource by id")
	}

	if _, ok := r.Tags[tag]; ok {
		return nil
	}

	if r.Tags == nil {
		r.Tags = make(map[string]*string)
	}

	r.Tags[tag] = &value
	genericResource := resources.GenericResource{
		Tags: r.Tags,
	}

	_, err = t.ResourcesClient.UpdateByID(context.Background(), id, genericResource)
	if err != nil {
		return errors.Wrap(err, "cannot update resource by id")
	}

	return err
}

// Execute executes action from p in resource data
func (t *Tagger) Execute(data *Resource, p rules.ActionItem) error {
	if val, ok := t.actionMap[p.GetType()]; ok {
		err := val(p, data)
		if err != nil {
			msg := fmt.Sprintf("Execute(action=%q) returned error %q", p.GetType(), err)
			return errors.New(msg)
		}
		return nil
	}
	log.Warnf("Unknown action type %s - ignoring", p.GetType())
	return nil
}

// Eval checks if condition p is satisfied on resource data
func (t *Tagger) Eval(data *Resource, p rules.ConditionItem) bool {
	if val, ok := t.condMap[p.GetType()]; ok {
		return val(p, data)
	}
	log.Warnf("Unknown condition type %s - ignoring", p.GetType())
	return false
}
