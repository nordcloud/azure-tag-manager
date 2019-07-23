package azure

import (
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
)

// TagChecker represents an Azure checker
type TagChecker struct {
	Session *session.AzureSession
}

// SameTagDifferentValue reprents a resource with a tag's value
type SameTagDifferentValue struct {
	Resource Resource
	Value    string
}

// CheckSameTagDifferentValue checks if resources in resources are tagged with the same tag but with different values. It returns a map of lists of such resources. The key to the list is tag key.
//TODO: make this differently
func (t TagChecker) CheckSameTagDifferentValue(resources []Resource) map[string][]SameTagDifferentValue {

	var (
		nonCompliant     = make(map[string][]SameTagDifferentValue)
		tagSeen          = make(map[string]SameTagDifferentValue)
		originalAppended = false
	)

	for _, resource := range resources {
		originalAppended = false
		for key, value := range resource.Tags {
			//if found
			if _, ok := tagSeen[key]; ok {
				// if already recorded but with a different value - this is what we are looking for
				if tagSeen[key].Value != *value {
					s := &SameTagDifferentValue{Resource: resource, Value: *value}
					nonCompliant[key] = append(nonCompliant[key], *s)
					if originalAppended == false {
						nonCompliant[key] = append(nonCompliant[key], tagSeen[key])
						originalAppended = true
					}
				}
			} else {
				tagSeen[key] = SameTagDifferentValue{Resource: resource, Value: *value}
			}
		}
	}
	return nonCompliant
}
