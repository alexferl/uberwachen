package loaders

import "github.com/alexferl/uberwachen/registries"

type Loader interface {
	Load(registry *registries.Handlers) error
}
