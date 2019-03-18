package azure

import (
	"context"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"bitbucket.org/nordcloud/tagmanager/internal/tagger"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/pkg/errors"
)

type ResourceGroupScanner struct {
	Session         *session.AzureSession
	ResourcesClient *resources.Client
	GroupsClient    *resources.GroupsClient
}

type Scanner interface {
	GetResources() ([]tagger.Resource, error)
	GetResourcesByResourceGroup(string) ([]tagger.Resource, error)
	GetGroups()
}

func String(v string) *string {
	return &v
}

func NewResourceGroupScanner(session *session.AzureSession) *ResourceGroupScanner {
	resClient := resources.NewClient(session.SubscriptionID)
	resClient.Authorizer = session.Authorizer
	grClient := resources.NewGroupsClient(session.SubscriptionID)
	grClient.Authorizer = session.Authorizer

	scanner := &ResourceGroupScanner{
		Session:         session,
		ResourcesClient: &resClient,
		GroupsClient:    &grClient,
	}

	return scanner
}

func (scanner ResourceGroupScanner) GetResources() ([]tagger.Resource, error) {

	groups, err := scanner.GetGroups()

	if err != nil {
		return nil, err
	}
	tab := make([]tagger.Resource, 0)
	// var err error
	for _, rg := range groups {
		for list, err := scanner.ResourcesClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.Next() {
			if err != nil {
				err = errors.Wrap(err, "got error while traverising resources list")
			}
			resource := list.Value()
			// fmt.Println(&rg)

			tab = append(tab, tagger.Resource{
				Platform: "azure", ID: *resource.ID, Name: resource.Name, Region: *resource.Location, Tags: resource.Tags, ResourceGroup: String(rg),
			})
		}
	}

	return tab, err
}

func (scanner ResourceGroupScanner) GetGroups() ([]string, error) {
	tab := make([]string, 0)
	var err error
	for list, err := scanner.GroupsClient.ListComplete(context.Background(), "", nil); list.NotDone(); err = list.Next() {
		if err != nil {
			err = errors.Wrap(err, "got error while traverising RG list")
		}
		rgName := *list.Value().Name
		tab = append(tab, rgName)
	}

	return tab, err
}

func (scanner ResourceGroupScanner) GetResourcesByResourceGroup(rg string) ([]tagger.Resource, error) {

	tab := make([]tagger.Resource, 0)
	var err error

	for list, err := scanner.ResourcesClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.Next() {
		if err != nil {
			return nil, errors.Wrap(err, "got error while traversing resources list")
		}

		resource := list.Value()
		tab = append(tab, tagger.Resource{
			Platform: "azure", ID: *resource.ID, Name: resource.Name, Region: *resource.Location, Tags: resource.Tags,
		})
	}

	return tab, err
}
