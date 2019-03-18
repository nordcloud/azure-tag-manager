package azure

import (
	"fmt"

	"bitbucket.org/nordcloud/tagmanager/internal/azure/session"
	"bitbucket.org/nordcloud/tagmanager/internal/tagger"
)

type TagChecker struct {
	Session *session.AzureSession
	Found   map[string]Found
	DryRun  bool
}

func (t TagChecker) CheckResourceGroup(resources []tagger.Resource) map[string][]tagger.Resource {
	var tagseen map[string]string
	var noncompliant map[string][]tagger.Resource
	noncompliant = make(map[string][]tagger.Resource)
	tagseen = make(map[string]string)
	// var noncompliance map[string][]string
	for _, resource := range resources {
		for key, value := range resource.Tags {
			if _, ok := tagseen[key]; ok {
				fmt.Println(key)
				if tagseen[key] != *value {
					fmt.Printf("Non compliance !! seen tag (%s=%s) != (%s=%s)\n", key, tagseen[key], key, *value)
					noncompliant[key] = append(noncompliant[key], resource)
					// fmt.Printf("Non compliant resources %v", resource)
				}
			} else {
				tagseen[key] = *value
			}
		}
	}
	return noncompliant
}

func NewAzureChecker() *TagChecker {
	checker := &TagChecker{}
	return checker
}
