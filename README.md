# Tag manager 

Currently the software works for Azure only. 

## Prerequisites

For Azure you need to create service principal.


## Donwload

You can download a binary for your architecture from https://bitbucket.org/nordcloud/tagmanager/downloads/ 

### Azure
Create service principal file:

```bash
az ad sp create-for-rbac --sdk-auth > my.auth
```

and export path to the authorizer:

```bash
export AZURE_AUTH_LOCATION=my.auth
```

## How it works ?

Tag rewriter accepts the payload where rules are defined. Each rule has a list of conditions and a lits of actions. If all conditions evaluate to true for a resource, all actions are executed. 

```json
{
  "dryrun": true,
  "rules":  [
    {
        "conditions": [
            {"type": "tagValue", "tag": "darek", "value" : "example"},
            {"type": "tagExists", "tag": "darek7"},
            {"type": "tagNotExists", "tag": "env"}
        ], 
        "actions": [
            {"type": "addTag", "tag": "mucha", "value": "zoo" },
            {"type": "addTag", "tag": "mucha3", "value": "zoo" }
        ]
      }
    ]
}
```

As of today, the tag rewriter accepts four kinds of conditions:

* `tagValue` - checks if a `tag` has a `value` set 
* `tagExists` - checks if a tag with key `tag` exists
* `tagNotExists` - same as above but negative
* `regionEqual` - checks if resource is in a region (or location in azure)
* `regionNotEqual` - same as above but negative


## Running 

`./tagmanager mapping.json` 

### Example run

```

➜  pantageusz git:(master) ✗ ./tagmanager mapping.json
INFO[0002] 👍  Conditions are true for (darek33) with ID = /subscriptions/6690b014-bdbd-4496-98ee-f2f255699f70/resourceGroups/darek/providers/Microsoft.Storage/storageAccounts/darek33 
INFO[0002]      🔥  DryRun Firing action addTag on resource /subscriptions/6690b014-bdbd-4496-98ee-f2f255699f70/resourceGroups/darek/providers/Microsoft.Storage/storageAccounts/darek33 
INFO[0002]      🔥  DryRun Firing action addTag on resource /subscriptions/6690b014-bdbd-4496-98ee-f2f255699f70/resourceGroups/darek/providers/Microsoft.Storage/storageAccounts/darek33 

```



## Licence 

Dariusz Dwornikowski, Nordcloud

