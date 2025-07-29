/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/pkg/config"

	"github.com/crossplane-contrib/provider-upjet-github/config/actions"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsenvironmentsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsenvironmentvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsorganizationpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsorganizationvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsrepositoryaccesslevel"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsrepositorypermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionssecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/actionsvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/branch"
	"github.com/crossplane-contrib/provider-upjet-github/config/branchdefault"
	"github.com/crossplane-contrib/provider-upjet-github/config/branchprotection"
	"github.com/crossplane-contrib/provider-upjet-github/config/branchprotectionv3"
	"github.com/crossplane-contrib/provider-upjet-github/config/emugroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/enterpriseorganization"
	"github.com/crossplane-contrib/provider-upjet-github/config/issuelabels"
	"github.com/crossplane-contrib/provider-upjet-github/config/membership"
	"github.com/crossplane-contrib/provider-upjet-github/config/organizationruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/repository"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositoryautolinkreference"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositorycollaborator"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositorydeploykey"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositoryenvironment"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositoryenvironmentdeploymentpolicy"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositoryfile"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositorypullrequest"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositoryruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/repositorywebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/team"
	"github.com/crossplane-contrib/provider-upjet-github/config/teammembers"
	"github.com/crossplane-contrib/provider-upjet-github/config/teammembership"
	"github.com/crossplane-contrib/provider-upjet-github/config/teamrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/teamsettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/teamsyncgroupmapping"
	"github.com/crossplane/upjet/pkg/registry/reference"
	"github.com/integrations/terraform-provider-github/v6/github"
)

const (
	resourcePrefix = "github"
	modulePath     = "github.com/crossplane-contrib/provider-upjet-github"
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
		actions.Configure,
		actionsenvironmentsecret.Configure,
		actionsenvironmentvariable.Configure,
		actionsorganizationpermissions.Configure,
		actionsorganizationsecret.Configure,
		actionsorganizationvariable.Configure,
		actionsrepositoryaccesslevel.Configure,
		actionsrepositorypermissions.Configure,
		actionssecret.Configure,
		actionsvariable.Configure,
		branch.Configure,
		branchdefault.Configure,
		branchprotection.Configure,
		branchprotectionv3.Configure,
		emugroupmapping.Configure,
		enterpriseorganization.Configure,
		issuelabels.Configure,
		membership.Configure,
		organizationruleset.Configure,
		repository.Configure,
		repositoryautolinkreference.Configure,
		repositorycollaborator.Configure,
		repositorydeploykey.Configure,
		repositoryenvironmentdeploymentpolicy.Configure,
		repositoryfile.Configure,
		repositorypullrequest.Configure,
		repositoryruleset.Configure,
		repositorywebhook.Configure,
		team.Configure,
		teammembers.Configure,
		teammembership.Configure,
		teamrepository.Configure,
		teamsettings.Configure,
		teamsyncgroupmapping.Configure,
		repositoryenvironment.Configure,
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
