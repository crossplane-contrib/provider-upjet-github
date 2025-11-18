/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"
)

const (
	resourcePrefix = "github"
	modulePath     = "github.com/crossplane-contrib/provider-upjet-github"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// resourceList returns the list of resources that have external
// name configured in the specified table.
func resourceList(t map[string]ujconfig.ExternalName) []string {
	l := make([]string, len(t))
	i := 0
	for n := range t {
		// Expected format is regex and we'd like to have exact matches.
		l[i] = n + "$"
		i++
	}
	return l
}
