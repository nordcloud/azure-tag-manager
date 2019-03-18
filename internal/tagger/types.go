package tagger

type CondFuncMap map[string]func(p map[string]string, data *Resource) bool
type ActionFuncMap map[string]func(p map[string]string, data *Resource) error

type Resource struct {
	Platform      string
	Name          *string
	Region        string
	ID            string
	Tags          map[string]*string
	ResourceGroup *string
}
