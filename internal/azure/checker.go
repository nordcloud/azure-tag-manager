package azure

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
)

type TagChecker struct {
	Session *session.AzureSession
	Found   map[string]Found
	DryRun  bool
}

func (t TagChecker) CheckResourceGroup(resources []Resource) map[string][]Resource {
	var (
		nonCompliant = make(map[string][]Resource)
		tagSeen      = make(map[string]string)
	)

	for _, resource := range resources {
		for key, value := range resource.Tags {
			if _, ok := tagSeen[key]; ok {
				fmt.Println(key)
				if tagSeen[key] != *value {
					fmt.Printf("Non compliance !! seen tag (%s=%s) != (%s=%s)\n", key, tagSeen[key], key, *value)
					nonCompliant[key] = append(nonCompliant[key], resource)
					// fmt.Printf("Non compliant resources %v", resource)
				}
			} else {
				tagSeen[key] = *value
			}
		}
	}

	return nonCompliant
}

func NewAzureChecker(s *session.AzureSession) *TagChecker {
	checker := TagChecker{
		Session: s,
	}

	return &checker
}
