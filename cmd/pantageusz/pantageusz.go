package main

import (
	"bitbucket.org/nordcloud/pantageusz/internal/azure"
	"bitbucket.org/nordcloud/pantageusz/internal/rules"

	log "github.com/sirupsen/logrus"
)

var payload = `{
	"rules":  [
	"dryrun": true,
	  {
		  "conditions": [
			  {"type": "tagValue", "tag": "darek", "value" : "dupa"},
			  {"type": "tagExists", "tag": "darek7"},
			  {"type": "tagNotExists", "tag": "env"}
		  ], 
		  "actions": [
			  {"type": "addTag", "tag": "mucha", "value": "zoo" },
			  {"type": "addTag", "tag": "mucha3", "value": "zoo" }
		  ],
	  }
	  ]
  }`

func main() {
	t, err := rules.NewRulesFromFile("mapper.json")

	if err != nil {
		log.WithError(err).Fatal("can't open rules file")
	}

	tagger, err := azure.NewAzureTagger(t)

	if err != nil {
		log.WithError(err).Fatal("Can't create tagger")
	}

	scanner := azure.SimpleScanner{Session: tagger.Session}
	res, err := scanner.GetResources()

	if err != nil {
		log.WithError(err).Fatalf("can't scan resources")
	}

	err = tagger.EvaluteRules(&res)
	if err != nil {
		log.WithError(err).Fatal("can't eval rules")
	}

	err = tagger.ExecuteActions()

	if err != nil {
		log.WithError(err).Fatal("can't exec actions")
	}
}
