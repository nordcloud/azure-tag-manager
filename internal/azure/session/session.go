package session

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"
)

// AzureSession stores subscription id and Authorized object
type AzureSession struct {
	SubscriptionID string
	Authorizer     autorest.Authorizer
}

// func readJSON(path string) (*map[string]interface{}, error) {
// 	data, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		log.Fatalf("failed to read file: %v", err)
// 	}

// 	contents := make(map[string]interface{})
// 	json.Unmarshal(data, &contents)
// 	return &contents, nil
// }

// NewFromFile creates new session from file kept in AZURE_AUTH_LOCATION.
func NewFromFile() (*AzureSession, error) {
	authorizer, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get initial session")
	}

	a, err := auth.GetSettingsFromFile()
	sess := AzureSession{
		SubscriptionID: a.GetSubscriptionID(),
		Authorizer:     authorizer,
	}

	return &sess, err
}
