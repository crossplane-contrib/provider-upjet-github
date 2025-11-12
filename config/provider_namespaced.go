/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	ujconfig "github.com/crossplane/upjet/v2/pkg/config"

	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsenvironmentsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsenvironmentvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsrepositoryaccesslevel"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsrepositorypermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionssecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branch"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branchdefault"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branchprotection"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branchprotectionv3"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/emugroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/enterpriseorganization"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/issuelabels"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/membership"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryautolinkreference"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorycollaborator"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorycustomproperty"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorydeploykey"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryenvironment"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryenvironmentdeploymentpolicy"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryfile"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorypullrequest"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorywebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/team"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teammembers"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teammembership"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teamrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teamsettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teamsyncgroupmapping"
	"github.com/crossplane/upjet/v2/pkg/registry/reference"
	"github.com/integrations/terraform-provider-github/v6/github"
)

// GetProviderNamespaced returns provider configuration
func GetProviderNamespaced(ctx context.Context) (*ujconfig.Provider, error) {

	pc := ujconfig.NewProvider([]byte(providerSchema), resourcePrefix, modulePath, []byte(providerMetadata),
		ujconfig.WithRootGroup("github.m.upbound.io"),
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
