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
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionshostedrunner"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationoidcsubjectclaimcustomizationtemplate"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationsecretrepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationsecretrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationvariablerepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationvariablerepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsorganizationworkflowpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsrepositoryaccesslevel"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsrepositoryoidcsubjectclaimcustomizationtemplate"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsrepositorypermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionssecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/actionsvariable"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/appinstallationrepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/appinstallationrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branch"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branchdefault"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branchprotection"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/branchprotectionv3"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/codespacesorganizationsecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/codespacesorganizationsecretrepositories"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/codespacessecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/codespacesusersecret"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/emugroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/enterpriseactionsrunnergroup"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/enterpriseactionsworkflowpermissions"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/enterpriseipallowlistentry"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/enterpriseorganization"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/enterprisesecurityanalysissettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/githubissue"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/issuelabel"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/issuelabels"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/membership"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationblock"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationrepositoryrole"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationrole"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationroleteam"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationroleuser"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/organizationwebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/release"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryautolinkreference"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorycollaborator"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorycollaborators"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorycustomproperty"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorydependabotsecurityupdates"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorydeploykey"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryenvironment"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryenvironmentdeploymentpolicy"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryfile"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorymilestone"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorypages"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorypullrequest"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryruleset"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorytopics"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositoryvulnerabilityalerts"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/repositorywebhook"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/team"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teammembers"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teammembership"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teamrepository"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teamsettings"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/teamsyncgroupmapping"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/usergpgkey"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/userinvitationaccepter"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/usersshkey"
	"github.com/crossplane-contrib/provider-upjet-github/config/cluster/workflowrepositorypermissions"
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
