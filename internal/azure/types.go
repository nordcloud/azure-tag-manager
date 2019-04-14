package azure

//Resource represents a generic resource with name, region, id, tags and resource group
type Resource struct {
	Platform      string
	Name          *string
	Region        string
	ID            string
	Tags          map[string]*string
	ResourceGroup *string
}

type condFuncMap map[string]func(p map[string]string, data *Resource) bool
type actionFuncMap map[string]func(p map[string]string, data *Resource) error
