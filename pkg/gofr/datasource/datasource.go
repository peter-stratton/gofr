package datasource

import "github.com/peter-stratton/gofr/pkg/gofr/config"

type Datasource interface {
	Register(config config.Config)
}

// Question is: is container aware exactly "Redis" is there or some opaque datasource. in the later case, how do we
// retrieve from context?
