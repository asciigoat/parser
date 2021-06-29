package shexp

import (
	"os"
)

// A Resolver is an object used by the Expander to resolve and store variables
type Resolver interface {
	Get(key string) string        // Gets value of a variable
	Set(key, value string) string // Sets a variable to a given value
}

type EnvResolver struct {
	extra map[string]string
}

func (r *EnvResolver) Get(key string) string {
	if r.extra != nil {
		if s, ok := r.extra[key]; ok {
			return s
		}
	}

	return os.Getenv(key)
}

func (r *EnvResolver) Set(key, value string) string {
	if r.extra == nil {
		r.extra = make(map[string]string, 1)
	}
	r.extra[key] = value

	return value
}

func (r *EnvResolver) Reset() {
	r.extra = make(map[string]string)
}
