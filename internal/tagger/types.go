package tagger

type CondFuncMap map[string]func(p map[string]string, data *TaggableResource) bool
type ActionFuncMap map[string]func(p map[string]string, data *TaggableResource) error

type TaggableResource struct {
	Platform      string
	Name          *string
	Region        string
	ID            string
	Tags          map[string]*string
	ResourceGroup *string
}
