package azure

import (
	"context"
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/nordcloud/azure-tag-manager/internal/azure/session"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources/resourcesapi"
	"github.com/pkg/errors"
)

type BackupEntry struct {
	ID   string             `json:"id"`
	Tags map[string]*string `json:"tags"`
}

type Restorer interface {
	Restore() error
}

type TagRestorer struct {
	Session         *session.AzureSession
	ResourcesClient resourcesapi.ClientAPI
	ReplaceTags     bool
	Backup          []BackupEntry
}

//NewBackupFromMatched make a file backup from the matching resources
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

func NewRestorerFromFile(filename string, s *session.AzureSession, replace bool) *TagRestorer {
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
		ReplaceTags:     replace,
	}
	return restorer
}
