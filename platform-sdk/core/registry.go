package core

import "fmt"

// PlatformFactory creates a Platform from config options.
type PlatformFactory func(opts map[string]any) (Platform, error)

var platformFactories = make(map[string]PlatformFactory)

// RegisterPlatform registers a platform factory by name.
func RegisterPlatform(name string, factory PlatformFactory) {
	platformFactories[name] = factory
}

// CreatePlatform creates a Platform instance by name with the given options.
func CreatePlatform(name string, opts map[string]any) (Platform, error) {
	f, ok := platformFactories[name]
	if !ok {
		available := make([]string, 0, len(platformFactories))
		for k := range platformFactories {
			available = append(available, k)
		}
		return nil, fmt.Errorf("unknown platform %q, available: %v", name, available)
	}
	return f(opts)
}

// ListPlatforms returns the names of all registered platforms.
func ListPlatforms() []string {
	names := make([]string, 0, len(platformFactories))
	for k := range platformFactories {
		names = append(names, k)
	}
	return names
}
