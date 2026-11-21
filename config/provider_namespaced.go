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
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionshostedrunner"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationoidcsubjectclaimcustomizationtemplate"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationsecretrepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationsecretrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationvariablerepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationvariablerepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsorganizationworkflowpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsrepositoryaccesslevel"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsrepositoryoidcsubjectclaimcustomizationtemplate"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsrepositorypermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionssecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/actionsvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/appinstallationrepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/appinstallationrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branch"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branchdefault"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branchprotection"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/branchprotectionv3"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/codespacesorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/codespacesorganizationsecretrepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/codespacessecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/codespacesusersecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/emugroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/enterpriseactionsrunnergroup"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/enterpriseactionsworkflowpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/enterpriseipallowlistentry"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/enterpriseorganization"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/enterprisesecurityanalysissettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/githubissue"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/issuelabel"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/issuelabels"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/membership"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationblock"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationrepositoryrole"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationrole"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationroleteam"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationroleuser"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/organizationwebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/release"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryautolinkreference"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorycollaborator"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorycollaborators"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorycustomproperty"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorydependabotsecurityupdates"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorydeploykey"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryenvironment"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryenvironmentdeploymentpolicy"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryfile"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorymilestone"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorypages"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorypullrequest"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorytopics"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositoryvulnerabilityalerts"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/repositorywebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/team"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teammembers"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teammembership"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teamrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teamsettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/teamsyncgroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/usergpgkey"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/userinvitationaccepter"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/usersshkey"
	"github.com/crossplane-contrib/provider-upjet-github/config/namespaced/workflowrepositorypermissions"
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
		ujconfig.WithTerraformProvider(github.NewProvider("dev", "none")()),
		ujconfig.WithDefaultResourceOptions(
			resourceConfigurator(),
		))

	for _, configure := range []func(provider *ujconfig.Provider){
		// add custom config functions
		actions.Configure,
		actionsenvironmentsecret.Configure,
		actionsenvironmentvariable.Configure,
		actionshostedrunner.Configure,
		actionsorganizationoidcsubjectclaimcustomizationtemplate.Configure,
		actionsorganizationpermissions.Configure,
		actionsorganizationsecret.Configure,
		actionsorganizationsecretrepositories.Configure,
		actionsorganizationsecretrepository.Configure,
		actionsorganizationvariable.Configure,
		actionsorganizationvariablerepositories.Configure,
		actionsorganizationvariablerepository.Configure,
		actionsorganizationworkflowpermissions.Configure,
		actionsrepositoryaccesslevel.Configure,
		actionsrepositoryoidcsubjectclaimcustomizationtemplate.Configure,
		actionsrepositorypermissions.Configure,
		actionssecret.Configure,
		actionsvariable.Configure,
		appinstallationrepositories.Configure,
		appinstallationrepository.Configure,
		branch.Configure,
		branchdefault.Configure,
		branchprotection.Configure,
		branchprotectionv3.Configure,
		codespacesorganizationsecret.Configure,
		codespacesorganizationsecretrepositories.Configure,
		codespacessecret.Configure,
		codespacesusersecret.Configure,
		emugroupmapping.Configure,
		enterpriseactionsrunnergroup.Configure,
		enterpriseactionsworkflowpermissions.Configure,
		enterpriseipallowlistentry.Configure,
		enterpriseorganization.Configure,
		enterprisesecurityanalysissettings.Configure,
		githubissue.Configure,
		issuelabel.Configure,
		issuelabels.Configure,
		membership.Configure,
		organizationblock.Configure,
		organizationrepositoryrole.Configure,
		organizationrole.Configure,
		organizationroleteam.Configure,
		organizationroleuser.Configure,
		organizationruleset.Configure,
		organizationwebhook.Configure,
		release.Configure,
		repository.Configure,
		repositoryautolinkreference.Configure,
		repositorycollaborator.Configure,
		repositorycollaborators.Configure,
		repositorycustomproperty.Configure,
		repositorydependabotsecurityupdates.Configure,
		repositorydeploykey.Configure,
		repositoryenvironmentdeploymentpolicy.Configure,
		repositoryfile.Configure,
		repositorymilestone.Configure,
		repositorypages.Configure,
		repositorypullrequest.Configure,
		repositoryruleset.Configure,
		repositorytopics.Configure,
		repositoryvulnerabilityalerts.Configure,
		repositorywebhook.Configure,
		team.Configure,
		teammembers.Configure,
		teammembership.Configure,
		teamrepository.Configure,
		teamsettings.Configure,
		teamsyncgroupmapping.Configure,
		repositoryenvironment.Configure,
		usergpgkey.Configure,
		userinvitationaccepter.Configure,
		usersshkey.Configure,
		workflowrepositorypermissions.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc, nil
}
