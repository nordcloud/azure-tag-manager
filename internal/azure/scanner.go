package azure

import (
	"context"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/pkg/errors"
)

type ResourceGroupScanner struct {
	Session         *session.AzureSession
	ResourcesClient *resources.Client
	GroupsClient    *resources.GroupsClient
}

type Scanner interface {
	GetResources() ([]Resource, error)
	GetResourcesByResourceGroup(string) ([]Resource, error)
	GetGroups() ([]string, error)
}

func String(v string) *string {
	return &v
}

func NewResourceGroupScanner(s *session.AzureSession) *ResourceGroupScanner {
	resClient := resources.NewClient(s.SubscriptionID)
	resClient.Authorizer = s.Authorizer

	grClient := resources.NewGroupsClient(s.SubscriptionID)
	grClient.Authorizer = s.Authorizer

	scanner := &ResourceGroupScanner{
		Session:         s,
		ResourcesClient: &resClient,
		GroupsClient:    &grClient,
	}

	return scanner
}

func (r ResourceGroupScanner) GetResources() ([]Resource, error) {
	groups, err := r.GetGroups()
	if err != nil {
		return nil, errors.Wrap(err, "could not obtain groups")
	}

	tab := make([]Resource, 0)
	for _, rg := range groups {
		for list, err := r.ResourcesClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.NextWithContext(context.Background()) {
			if err != nil {
				return nil, errors.Wrap(err, "got error while traversing resources list")
			}

			resource := list.Value()
			tab = append(tab, Resource{
				Platform:      "azure",
				ID:            *resource.ID,
				Name:          resource.Name,
				Region:        *resource.Location,
				Tags:          resource.Tags,
				ResourceGroup: String(rg),
			})
		}
	}

	return tab, nil
}

func (r ResourceGroupScanner) GetGroups() ([]string, error) {
	tab := make([]string, 0)
	for list, err := r.GroupsClient.ListComplete(context.Background(), "", nil); list.NotDone(); err = list.NextWithContext(context.Background()) {
		if err != nil {
			return nil, errors.Wrap(err, "got error while traverising RG list")
		}

		rgName := *list.Value().Name
		tab = append(tab, rgName)
	}

	return tab, nil
}

func (r ResourceGroupScanner) GetResourcesByResourceGroup(rg string) ([]Resource, error) {
	tab := make([]Resource, 0)
	for list, err := r.ResourcesClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.NextWithContext(context.Background()) {
		if err != nil {
			return nil, errors.Wrap(err, "got error while traversing resources list")
		}

		resource := list.Value()
		tab = append(tab, Resource{
			Platform: "azure",
			ID:       *resource.ID,
			Name:     resource.Name,
			Region:   *resource.Location,
			Tags:     resource.Tags,
		})
	}

	return tab, nil
}
