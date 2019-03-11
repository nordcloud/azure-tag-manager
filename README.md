# Tag manager 

Currently the software works for Azure only. 

## Prerequisites

For Azure you need to create service principal.

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
        ],
        "dryrun": true
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

`./pantageusz mapping.json` 

## Why the name? 

[Pan Tadeusz](https://en.wikipedia.org/wiki/Pan_Tadeusz) by Adam Mickiewicz is a national poilish epic poem. `Pan Tadeusz` means `Sir Thadeus` and Pan `Tag`eusz is a pun on this (funny word play). 

## Licence 

Dariusz Dwornikowski, Nordcloud

