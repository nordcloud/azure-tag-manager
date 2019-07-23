package azure

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/nordcloud/azure-tag-manager/internal/azure/rules"
	"github.com/nordcloud/azure-tag-manager/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	twoRulesWant = rules.TagRules{Rules: []rules.Rule{
		{Name: "name", Conditions: []rules.ConditionItem{
			{"type": "tagEqual", "tag": "test", "value": "test"},
			{"type": "tagExists", "tag": "test"},
		},
			Actions: []rules.ActionItem{
				{"type": "addTag", "tag": "test2", "value": "test2"},
			},
		},
	}}

	deleteTag = rules.TagRules{Rules: []rules.Rule{
		{Name: "name", Conditions: []rules.ConditionItem{
			{"type": "tagEqual", "tag": "test2", "value": "test2"},
		},
			Actions: []rules.ActionItem{
				{"type": "delTag", "tag": "test3"},
			},
		},
	}}

	deleteAllTags = rules.TagRules{Rules: []rules.Rule{
		{Name: "name", Conditions: []rules.ConditionItem{
			{"type": "tagEqual", "tag": "test2", "value": "test2"},
		},
			Actions: []rules.ActionItem{
				{"type": "cleanTags"},
			},
		},
	}}
)

var testResources = []Resource{
	{ID: "1", Region: "westeurope", Tags: map[string]*string{"test": String("test")}, ResourceGroup: String("test"), Name: String("name")},
	{ID: "2", Region: "westeurope", Tags: map[string]*string{"test2": String("test2"), "test3": String("test3")}, ResourceGroup: String("te3st"), Name: String("name2")},
	{ID: "3", Region: "easteurope", Tags: map[string]*string{"test-region": String("other"), "othertest": String("test56")}, ResourceGroup: String("rg2"), Name: String("name3")},
}

func TestTagger_ExecuteActions(t *testing.T) {
	mockClient := new(mocks.ClientAPI)
	mockClient.On("GetByID", context.Background(), "1").Return(resources.GenericResource{ID: String("1"), Location: String("weseurope"), Name: String("test")}, nil)
	mockClient.On("GetByID", context.Background(), "2").Return(resources.GenericResource{ID: String("2"), Location: String("weseurope"), Name: String("name2")}, nil)
	mockClient.On("UpdateByID", context.Background(), "1", resources.GenericResource{Tags: map[string]*string{"test2": String("test2")}}).Return(resources.UpdateByIDFuture{}, nil)
	mockClient.On("UpdateByID", context.Background(), "2", resources.GenericResource{Tags: map[string]*string{"test2": String("test2")}}).Return(resources.UpdateByIDFuture{}, nil)

	t.Run("Test addTag on resource", func(t *testing.T) {
		tagger := Tagger{
			ResourcesClient: mockClient,
			Rules:           twoRulesWant,
			Matched:         make(map[string]Matched),
		}
		tagger.InitActionMap()
		tagger.InitCondMap()
		tagger.EvaluateRules(testResources)
		assert.Contains(t, tagger.Matched, "1")
		ael, err := tagger.ExecuteActions()
		assert.Nil(t, err)
		assert.Len(t, ael, 1)
	})

	t.Run("Test delTag on resource", func(t *testing.T) {
		tagger := Tagger{
			ResourcesClient: mockClient,
			Rules:           deleteTag,
			Matched:         make(map[string]Matched),
		}
		tagger.InitActionMap()
		tagger.InitCondMap()
		tagger.EvaluateRules(testResources)
		assert.Contains(t, tagger.Matched, "2")
		ael, err := tagger.ExecuteActions()
		assert.Nil(t, err)
		assert.Len(t, ael, 1)
	})

	mockClient.On("UpdateByID", context.Background(), "2", resources.GenericResource{Tags: map[string]*string{}}).Return(resources.UpdateByIDFuture{}, nil)

	t.Run("Test delete all tags on resource", func(t *testing.T) {
		tagger := Tagger{
			ResourcesClient: mockClient,
			Rules:           deleteAllTags,
			Matched:         make(map[string]Matched),
		}
		tagger.InitActionMap()
		tagger.InitCondMap()
		tagger.EvaluateRules(testResources)
		assert.Contains(t, tagger.Matched, "2")
		ael, err := tagger.ExecuteActions()
		assert.Nil(t, err)
		assert.Len(t, ael, 1)
	})
}
