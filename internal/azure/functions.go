package azure

import (
	"context"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/rules"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (t Tagger) deleteAllTags(id string) error {
	client := resources.NewClient(t.Session.SubscriptionID)
	client.Authorizer = t.Session.Authorizer

	_, err := client.GetByID(context.Background(), id)
	if err != nil {
		return errors.Wrap(err, "cannot get resource by id")
	}

	genericResource := resources.GenericResource{
		Tags: make(map[string]*string),
	}

	_, err = client.UpdateByID(context.Background(), id, genericResource)
	if err != nil {
		return errors.Wrap(err, "cannot update resource by id")
	}

	return err
}

func (t Tagger) deleteTag(id, tag string) error {
	client := resources.NewClient(t.Session.SubscriptionID)
	client.Authorizer = t.Session.Authorizer

	r, err := client.GetByID(context.Background(), id)
	if err != nil {
		return errors.Wrap(err, "cannot get resource by id")
	}

	if _, ok := r.Tags[tag]; !ok {
		return nil
	}

	delete(r.Tags, tag)
	genericResource := resources.GenericResource{
		Tags: r.Tags,
	}

	_, err = client.UpdateByID(context.Background(), id, genericResource)
	if err != nil {
		return errors.Wrap(err, "cannot update resource by id")
	}

	return err
}

func (t Tagger) createOrUpdateTag(id, tag, value string) error {
	client := resources.NewClient(t.Session.SubscriptionID)
	client.Authorizer = t.Session.Authorizer

	r, err := client.GetByID(context.Background(), id)
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

	_, err = client.UpdateByID(context.Background(), id, genericResource)
	if err != nil {
		return errors.Wrap(err, "cannot update resource by id")
	}

	return err
}

func (t *Tagger) Execute(data *Resource, p rules.ActionItem) error {
	if val, ok := t.actionMap[p.GetType()]; ok {
		return val(p, data)
	}

	log.Warnf("Unknown action type %s", p.GetType())
	return nil
}

func (t *Tagger) Eval(data *Resource, p rules.ConditionItem) bool {
	if val, ok := t.condMap[p.GetType()]; ok {
		return val(p, data)
	}

	log.Warnf("Unknown condition type %s", p.GetType())
	return false
}
