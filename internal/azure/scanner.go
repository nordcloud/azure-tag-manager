package azure

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
	"github.com/pkg/errors"
)

// ResourceGroupScanner represents resource group scanner that scans all resources in a resource group
type ResourceGroupScanner struct {
	Session         *session.AzureSession
	ResourcesClient *resources.Client
	GroupsClient    *resources.GroupsClient
}

// Scanner represents generic scanner of Azure resource groups
type Scanner interface {
	GetResources() ([]Resource, error)
	GetResourcesByResourceGroup(string) ([]Resource, error)
	GetGroups() ([]string, error)
	GetResourceGroupTags(string) (map[string]*string, error)
}

// String converts string v to the string pointer
func String(v string) *string {
	return &v
}

// GetResourceGroupTags returns a map of key value tags of a reource group rg
func (r ResourceGroupScanner) GetResourceGroupTags(rg string) (map[string]*string, error) {
	result, err := r.GroupsClient.Get(context.Background(), rg)
	if err != nil {
		return nil, errors.Wrapf(err, "GetResourceGroupTags(rg=%s): Get() failed", rg)
	}
	return result.Tags, nil
}

// NewResourceGroupScanner creates ResourceGroupScanner with Azure Serssion s
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

// ScanResourceGroup returns a list of resources and their tags from a resource group rg
func (r ResourceGroupScanner) ScanResourceGroup(rg string) []Resource {
	tab := make([]Resource, 0)
	for list, err := r.ResourcesClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.NextWithContext(context.Background()) {
		if err != nil {
			log.Fatal(err)
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
	return tab
}

// GetResources retruns list of resources in resource group
func (r ResourceGroupScanner) GetResources() ([]Resource, error) {
	var wg sync.WaitGroup

	groups, err := r.GetGroups()
	if err != nil {
		return nil, errors.Wrap(err, "GetResources(): GetGroups() failed")
	}

	tab := make([]Resource, 0)
	out := make(chan []Resource)
	for _, rg := range groups {
		wg.Add(1)
		go func(rg string) {
			defer wg.Done()
			out <- r.ScanResourceGroup(rg)
		}(rg)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	for s := range out {
		tab = append(tab, s...)
	}

	return tab, nil
}

// GetGroups returns list of resource groups in a subscription
func (r ResourceGroupScanner) GetGroups() ([]string, error) {
	tab := make([]string, 0)
	for list, err := r.GroupsClient.ListComplete(context.Background(), "", nil); list.NotDone(); err = list.NextWithContext(context.Background()) {
		if err != nil {
			return nil, errors.Wrap(err, "GetGroups(): GroupsClient.ListComplete failed")
		}
		rgName := *list.Value().Name
		tab = append(tab, rgName)
	}
	return tab, nil
}

// GetResourcesByResourceGroup returns resources in a resource group rg
func (r ResourceGroupScanner) GetResourcesByResourceGroup(rg string) ([]Resource, error) {
	tab := make([]Resource, 0)
	for list, err := r.ResourcesClient.ListByResourceGroupComplete(context.Background(), rg, "", "", nil); list.NotDone(); err = list.NextWithContext(context.Background()) {
		if err != nil {
			return nil, errors.Wrapf(err, "GetResourcesByResourceGroup(rg=%q): ListByResourceGroupComplete() failed", rg)
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
