/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	"github.com/coopnorge/provider-github/config/actionssecret"
	"github.com/coopnorge/provider-github/config/actionsvariable"
	"github.com/coopnorge/provider-github/config/branch"
	"github.com/coopnorge/provider-github/config/branchprotection"
	"github.com/coopnorge/provider-github/config/defaultbranch"
	"github.com/coopnorge/provider-github/config/deploykey"
	"github.com/coopnorge/provider-github/config/emugroupmapping"
	"github.com/coopnorge/provider-github/config/membership"
	"github.com/coopnorge/provider-github/config/organization"
	"github.com/coopnorge/provider-github/config/organizationruleset"
	"github.com/coopnorge/provider-github/config/pullrequest"
	"github.com/coopnorge/provider-github/config/repository"
	"github.com/coopnorge/provider-github/config/repositoryfile"
	"github.com/coopnorge/provider-github/config/repositorywebhook"
	"github.com/coopnorge/provider-github/config/team"
	"github.com/coopnorge/provider-github/config/teammembership"
	"github.com/coopnorge/provider-github/config/teamrepository"
	"github.com/coopnorge/provider-github/config/teamsettings"
	"github.com/coopnorge/provider-github/config/teamsyncgroupmapping"
	ujconfig "github.com/crossplane/upjet/pkg/config"
	"github.com/crossplane/upjet/pkg/registry/reference"
	"github.com/integrations/terraform-provider-github/v6/github"
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
func GetProvider(ctx context.Context) (*ujconfig.Provider, error) {

	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("github.upbound.io"),
		ujconfig.WithShortName("github"),
		ujconfig.WithIncludeList(resourceList(cliReconciledExternalNameConfigs)),
		ujconfig.WithTerraformPluginSDKIncludeList(resourceList(terraformPluginSDKExternalNameConfigs)),
		ujconfig.WithFeaturesPackage("internal/features"),
		ujconfig.WithReferenceInjectors([]ujconfig.ReferenceInjector{reference.NewInjector(modulePath)}),
		ujconfig.WithTerraformProvider(github.Provider()),
		ujconfig.WithDefaultResourceOptions(
			resourceConfigurator(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		repository.Configure,
		branch.Configure,
		deploykey.Configure,
		repositoryfile.Configure,
		pullrequest.Configure,
		team.Configure,
		emugroupmapping.Configure,
		teammembership.Configure,
		teamrepository.Configure,
		defaultbranch.Configure,
		branchprotection.Configure,
		repositorywebhook.Configure,
		actionssecret.Configure,
		actionsvariable.Configure,
		organization.Configure,
		organizationruleset.Configure,
		membership.Configure,
		teamsettings.Configure,
		teamsyncgroupmapping.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc, nil
}

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
