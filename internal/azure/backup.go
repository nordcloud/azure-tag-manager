package azure

import (
	"context"
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources/resourcesapi"
	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
	"github.com/pkg/errors"
)

// BackupEntry represents one resource tags backup
type BackupEntry struct {
	ID   string             `json:"id"`
	Tags map[string]*string `json:"tags"`
}

// Restorer provides interface for restorers
type Restorer interface {
	Restore() error
}

// TagRestorer represents a restorer of Azure tags from backup
type TagRestorer struct {
	Session         *session.AzureSession  // session to connect to Azure
	ResourcesClient resourcesapi.ClientAPI // client to the resources API
	Backup          []BackupEntry          // list of backup entries
}

//NewBackupFromMatched makes a file backup from the resources in matched to a json file in directory
func NewBackupFromMatched(matched map[string]Matched, directory string) string {
	var backup []BackupEntry

	for ID, matched := range matched {
		entry := &BackupEntry{
			ID:   ID,
			Tags: matched.Resource.Tags,
		}
		backup = append(backup, *entry)
	}
	tmpfile, err := ioutil.TempFile(directory, "tagmanager.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpfile.Close()

	jsonBackup, err := json.Marshal(backup)

	if _, err := tmpfile.Write(jsonBackup); err != nil {
		tmpfile.Close()
		log.Fatal(err)
	}

	return tmpfile.Name()
}

// Restore restores tags from a backup file provided in TagRestorer
func (t TagRestorer) Restore() error {
	for _, backupEntry := range t.Backup {
		log.Infof("Restoring tags for [%s]\n", backupEntry.ID)
		_, err := t.ResourcesClient.GetByID(context.Background(), backupEntry.ID)

		if err != nil {
			return errors.Wrap(err, "cannot get resource by id")
		}

		genericResource := resources.GenericResource{
			Tags: backupEntry.Tags,
		}
		_, err = t.ResourcesClient.UpdateByID(context.Background(), backupEntry.ID, genericResource)
		if err != nil {
			return errors.Wrapf(err, "cannot update resource %s by id", backupEntry.ID)
		}
	}
	return nil
}

// NewRestorerFromFile creates a TagRestorer, which will restore tag backup from filename
func NewRestorerFromFile(filename string, s *session.AzureSession) *TagRestorer {
	resClient := resources.NewClient(s.SubscriptionID)
	resClient.Authorizer = s.Authorizer

	var backup []BackupEntry
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	byt := []byte(dat)
	if err := json.Unmarshal(byt, &backup); err != nil {
		log.Fatal(err)
	}

	restorer := &TagRestorer{
		Session:         s,
		ResourcesClient: &resClient,
		Backup:          backup,
	}
	return restorer
}
