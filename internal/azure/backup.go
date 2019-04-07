package azure

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type BackupEntry struct {
	ID   string             `json:"id"`
	Tags map[string]*string `json:"tags"`
}

func NewBackupFromFound(found map[string]Found, directory string) string {
	var backup []BackupEntry

	for ID, found := range found {
		entry := &BackupEntry{
			ID:   ID,
			Tags: found.Resource.Tags,
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
