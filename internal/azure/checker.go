package azure

import (
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
)

type TagChecker struct {
	Session *session.AzureSession
}

type SameTagDifferentValue struct {
	Resource Resource
	Value    string
}

func (t TagChecker) CheckSameTagDifferentValue(resources []Resource) map[string][]SameTagDifferentValue {

	var (
		nonCompliant     = make(map[string][]SameTagDifferentValue)
		tagSeen          = make(map[string]SameTagDifferentValue)
		originalAppended = false
	)

	for _, resource := range resources {
		for key, value := range resource.Tags {
			originalAppended = false
			//if exists
			if _, ok := tagSeen[key]; ok {
				// if already recorded with different value
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

//NewTagChecker 
func NewTagChecker(s *session.AzureSession) *TagChecker {
	checker := TagChecker{
		Session: s,
	}

	return &checker
}
