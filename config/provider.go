/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	"github.com/coopnorge/provider-github/config/actionssecret"
	"github.com/coopnorge/provider-github/config/branch"
	"github.com/coopnorge/provider-github/config/branchprotection"
	"github.com/coopnorge/provider-github/config/defaultbranch"
	"github.com/coopnorge/provider-github/config/repository"
	"github.com/coopnorge/provider-github/config/repositoryfile"
	"github.com/coopnorge/provider-github/config/team"
	"github.com/coopnorge/provider-github/config/teamrepository"
	"github.com/coopnorge/provider-github/config/pullrequest"
	ujconfig "github.com/upbound/upjet/pkg/config"
)

const (
	resourcePrefix = "github"
	modulePath     = "github.com/coopnorge/provider-github"
)

//go:embed schema.json
var providerSchema string

//go:embed provider-metadata.yaml
var providerMetadata string

// GetProvider returns provider configuration
func GetProvider() *ujconfig.Provider {
	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithIncludeList(ExternalNameConfigured()),
		ujconfig.WithDefaultResourceOptions(
			ExternalNameConfigurations(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		repository.Configure,
		branch.Configure,
		repositoryfile.Configure,
		pullrequest.Configure,
		team.Configure,
		teamrepository.Configure,
		defaultbranch.Configure,
		branchprotection.Configure,
		repositoryfile.Configure,
		actionssecret.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}
