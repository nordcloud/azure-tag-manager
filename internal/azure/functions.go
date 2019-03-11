package azure

import (
	"context"

	"bitbucket.org/nordcloud/tagmanager/internal/rules"

	"bitbucket.org/nordcloud/tagmanager/internal/tagger"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/pkg/errors"
)

func (t Tagger) createOrUpdateTag(id, tag, value string) error {

	client := resources.NewClient(t.Session.SubscriptionID)
	client.Authorizer = t.Session.Authorizer
	r, err := client.GetByID(context.Background(), id)

	if err != nil {
		return errors.Wrap(err, "cannot get resource by id")
	}
	if _, ok := r.Tags[tag]; ok {
		// log.Warnf("Tag %s already exists on resource %s", tag, id)
		return nil
	}

	r.Tags[tag] = &value
	genericResource := resources.GenericResource{
		Tags: r.Tags,
	}
	_, err = client.UpdateByID(context.Background(), id, genericResource)

	if err != nil {
		return errors.Wrap(err, "cannot get resource by id")
	}
	return err
}

func (t *Tagger) Execute(data *tagger.TaggableResource, p rules.ActionItem) error {
	if val, ok := t.actionMap[p.GetType()]; ok {
		return val(p, data)
	}
	return nil
}

func (t *Tagger) Eval(data *tagger.TaggableResource, p rules.ConditionItem) bool {

	if val, ok := t.condMap[p.GetType()]; ok {
		return val(p, data)
	}
	return false
}
