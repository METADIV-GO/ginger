package ginger

import (
	"os"
	"slices"
)

/*
Env returns the environment variable value by the key.
The key will be stored in the engine.
*/
func Env(key string) string {
	if !slices.Contains(Engine.EnvironmentKeys, key) {
		Engine.EnvironmentKeys = append(Engine.EnvironmentKeys, key)
	}
	return os.Getenv(key)
}
