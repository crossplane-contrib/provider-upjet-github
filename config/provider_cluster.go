/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsenvironmentsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsenvironmentvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsrepositoryaccesslevel"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsrepositorypermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionssecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branch"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branchdefault"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branchprotection"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branchprotectionv3"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/emugroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/enterpriseorganization"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/issuelabels"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/membership"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryautolinkreference"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorycollaborator"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorycustomproperty"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorydeploykey"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryenvironment"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryenvironmentdeploymentpolicy"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryfile"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorypullrequest"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorywebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/team"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teammembers"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teammembership"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teamrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teamsettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teamsyncgroupmapping"
	"github.com/crossplane/upjet/v2/pkg/registry/reference"
	"github.com/integrations/terraform-provider-github/v6/github"
)

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
		repositorycustomproperty.Configure,
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
