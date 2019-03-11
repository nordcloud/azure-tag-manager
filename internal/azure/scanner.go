package azure

import (
	"context"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"bitbucket.org/nordcloud/tagmanager/internal/tagger"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/pkg/errors"
)

type SimpleScanner struct {
	Session *session.AzureSession
}

type ResourceGroupScanner struct {
	Session *session.AzureSession
}

type Scanner interface {
	GetResources() ([]tagger.TaggableResource, error)
}

func String(v string) *string {
	return &v
}

func (scanner ResourceGroupScanner) GetResources() ([]tagger.TaggableResource, error) {
	grClient := resources.NewClient(scanner.Session.SubscriptionID)
	grClient.Authorizer = scanner.Session.Authorizer
	groups, err := getGroups(scanner.Session)

	if err != nil {
		return nil, err
	}
	tab := make([]tagger.TaggableResource, 0)
	// var err error
	for _, rg := range groups {
		for list, err := grClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.Next() {
			if err != nil {
				err = errors.Wrap(err, "got error while traverising resources list")
			}
			resource := list.Value()
			// fmt.Println(&rg)

			tab = append(tab, tagger.TaggableResource{
				Platform: "azure", ID: *resource.ID, Name: resource.Name, Region: *resource.Location, Tags: resource.Tags, ResourceGroup: String(rg),
			})
		}
	}

	return tab, err
}

func getGroups(sess *session.AzureSession) ([]string, error) {
	grClient := resources.NewGroupsClient(sess.SubscriptionID)
	grClient.Authorizer = sess.Authorizer
	tab := make([]string, 0)
	var err error
	for list, err := grClient.ListComplete(context.Background(), "", nil); list.NotDone(); err = list.Next() {
		if err != nil {
			err = errors.Wrap(err, "got error while traverising RG list")
		}
		rgName := *list.Value().Name
		tab = append(tab, rgName)
	}

	return tab, err
}

func (scanner SimpleScanner) GetResources() ([]tagger.TaggableResource, error) {
	grClient := resources.NewClient(scanner.Session.SubscriptionID)
	grClient.Authorizer = scanner.Session.Authorizer

	tab := make([]tagger.TaggableResource, 0)
	var err error
	for list, err := grClient.ListComplete(context.Background(), "", "", nil); list.NotDone(); err = list.Next() {
		if err != nil {
			err = errors.Wrap(err, "got error while traverising resources list")
		}
		resource := list.Value()
		tab = append(tab, tagger.TaggableResource{
			Platform: "azure", ID: *resource.ID, Name: resource.Name, Region: *resource.Location, Tags: resource.Tags,
		})
	}

	return tab, err
}
