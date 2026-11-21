// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	actionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/actionssecret"
	actionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/actionsvariable"
	environmentsecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/environmentsecret"
	environmentvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/environmentvariable"
	hostedrunner "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/hostedrunner"
	organizationactionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationactionssecret"
	organizationactionssecretrepository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationactionssecretrepository"
	organizationactionssecretrepositoryset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationactionssecretrepositoryset"
	organizationactionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationactionsvariable"
	organizationactionsvariablerepository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationactionsvariablerepository"
	organizationactionsvariablerepositoryset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationactionsvariablerepositoryset"
	organizationoidcsubjectclaimcustomizationtemplate "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationoidcsubjectclaimcustomizationtemplate"
	organizationpermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationpermissions"
	organizationworkflowpermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/organizationworkflowpermissions"
	repositoryaccesslevel "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/repositoryaccesslevel"
	repositoryoidcsubjectclaimcustomizationtemplate "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/repositoryoidcsubjectclaimcustomizationtemplate"
	repositorypermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/repositorypermissions"
	repositoryworkflowpermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/repositoryworkflowpermissions"
	runnergroup "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/actions/runnergroup"
	appinstallationrepositories "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/app/appinstallationrepositories"
	appinstallationrepositoryassociation "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/app/appinstallationrepositoryassociation"
	codespacessecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/codespaces/codespacessecret"
	organizationcodespacessecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/codespaces/organizationcodespacessecret"
	organizationcodespacessecretrepositoryset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/codespaces/organizationcodespacessecretrepositoryset"
	usercodespacessecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/codespaces/usercodespacessecret"
	actionsrunnergroup "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/enterprise/actionsrunnergroup"
	actionsworkflowpermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/enterprise/actionsworkflowpermissions"
	ipallowlistentry "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/enterprise/ipallowlistentry"
	organization "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/enterprise/organization"
	organizationruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/enterprise/organizationruleset"
	securityanalysissettings "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/enterprise/securityanalysissettings"
	organizationblock "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/organization/organizationblock"
	organizationrepositoryrole "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/organization/organizationrepositoryrole"
	organizationrole "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/organization/organizationrole"
	organizationwebhook "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/organization/organizationwebhook"
	providerconfig "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/providerconfig"
	branch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/branch"
	branchprotection "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/branchprotection"
	branchprotectionv3 "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/branchprotectionv3"
	defaultbranch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/defaultbranch"
	deploykey "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/deploykey"
	environment "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/environment"
	environmentdeploymentpolicy "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/environmentdeploymentpolicy"
	githubissue "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/githubissue"
	issuelabels "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/issuelabels"
	pullrequest "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/pullrequest"
	release "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/release"
	repository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repository"
	repositoryautolinkreference "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositoryautolinkreference"
	repositorycollaborator "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorycollaborator"
	repositorycollaboratorset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorycollaboratorset"
	repositorycustomproperty "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorycustomproperty"
	repositorydependabotsecurityupdates "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorydependabotsecurityupdates"
	repositoryfile "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositoryfile"
	repositorylabel "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorylabel"
	repositorymilestone "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorymilestone"
	repositorypages "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorypages"
	repositoryruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositoryruleset"
	repositorytopics "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorytopics"
	repositoryvulnerabilityalerts "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositoryvulnerabilityalerts"
	repositorywebhook "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/repo/repositorywebhook"
	emugroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/emugroupmapping"
	members "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/members"
	organizationroleteam "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/organizationroleteam"
	team "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/team"
	teammembership "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/teammembership"
	teamrepository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/teamrepository"
	teamsettings "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/teamsettings"
	teamsyncgroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/team/teamsyncgroupmapping"
	gpgkey "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/user/gpgkey"
	invitationaccepter "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/user/invitationaccepter"
	membership "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/user/membership"
	organizationroleuser "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/user/organizationroleuser"
	sshkey "github.com/crossplane-contrib/provider-upjet-github/internal/controller/cluster/user/sshkey"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		actionssecret.Setup,
		actionsvariable.Setup,
		environmentsecret.Setup,
		environmentvariable.Setup,
		hostedrunner.Setup,
		organizationactionssecret.Setup,
		organizationactionssecretrepository.Setup,
		organizationactionssecretrepositoryset.Setup,
		organizationactionsvariable.Setup,
		organizationactionsvariablerepository.Setup,
		organizationactionsvariablerepositoryset.Setup,
		organizationoidcsubjectclaimcustomizationtemplate.Setup,
		organizationpermissions.Setup,
		organizationworkflowpermissions.Setup,
		repositoryaccesslevel.Setup,
		repositoryoidcsubjectclaimcustomizationtemplate.Setup,
		repositorypermissions.Setup,
		repositoryworkflowpermissions.Setup,
		runnergroup.Setup,
		appinstallationrepositories.Setup,
		appinstallationrepositoryassociation.Setup,
		codespacessecret.Setup,
		organizationcodespacessecret.Setup,
		organizationcodespacessecretrepositoryset.Setup,
		usercodespacessecret.Setup,
		actionsrunnergroup.Setup,
		actionsworkflowpermissions.Setup,
		ipallowlistentry.Setup,
		organization.Setup,
		organizationruleset.Setup,
		securityanalysissettings.Setup,
		organizationblock.Setup,
		organizationrepositoryrole.Setup,
		organizationrole.Setup,
		organizationwebhook.Setup,
		providerconfig.Setup,
		branch.Setup,
		branchprotection.Setup,
		branchprotectionv3.Setup,
		defaultbranch.Setup,
		deploykey.Setup,
		environment.Setup,
		environmentdeploymentpolicy.Setup,
		githubissue.Setup,
		issuelabels.Setup,
		pullrequest.Setup,
		release.Setup,
		repository.Setup,
		repositoryautolinkreference.Setup,
		repositorycollaborator.Setup,
		repositorycollaboratorset.Setup,
		repositorycustomproperty.Setup,
		repositorydependabotsecurityupdates.Setup,
		repositoryfile.Setup,
		repositorylabel.Setup,
		repositorymilestone.Setup,
		repositorypages.Setup,
		repositoryruleset.Setup,
		repositorytopics.Setup,
		repositoryvulnerabilityalerts.Setup,
		repositorywebhook.Setup,
		emugroupmapping.Setup,
		members.Setup,
		organizationroleteam.Setup,
		team.Setup,
		teammembership.Setup,
		teamrepository.Setup,
		teamsettings.Setup,
		teamsyncgroupmapping.Setup,
		gpgkey.Setup,
		invitationaccepter.Setup,
		membership.Setup,
		organizationroleuser.Setup,
		sshkey.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupGated creates all controllers with the supplied logger and adds them to
// the supplied manager gated.
func SetupGated(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		actionssecret.SetupGated,
		actionsvariable.SetupGated,
		environmentsecret.SetupGated,
		environmentvariable.SetupGated,
		hostedrunner.SetupGated,
		organizationactionssecret.SetupGated,
		organizationactionssecretrepository.SetupGated,
		organizationactionssecretrepositoryset.SetupGated,
		organizationactionsvariable.SetupGated,
		organizationactionsvariablerepository.SetupGated,
		organizationactionsvariablerepositoryset.SetupGated,
		organizationoidcsubjectclaimcustomizationtemplate.SetupGated,
		organizationpermissions.SetupGated,
		organizationworkflowpermissions.SetupGated,
		repositoryaccesslevel.SetupGated,
		repositoryoidcsubjectclaimcustomizationtemplate.SetupGated,
		repositorypermissions.SetupGated,
		repositoryworkflowpermissions.SetupGated,
		runnergroup.SetupGated,
		appinstallationrepositories.SetupGated,
		appinstallationrepositoryassociation.SetupGated,
		codespacessecret.SetupGated,
		organizationcodespacessecret.SetupGated,
		organizationcodespacessecretrepositoryset.SetupGated,
		usercodespacessecret.SetupGated,
		actionsrunnergroup.SetupGated,
		actionsworkflowpermissions.SetupGated,
		ipallowlistentry.SetupGated,
		organization.SetupGated,
		organizationruleset.SetupGated,
		securityanalysissettings.SetupGated,
		organizationblock.SetupGated,
		organizationrepositoryrole.SetupGated,
		organizationrole.SetupGated,
		organizationwebhook.SetupGated,
		providerconfig.SetupGated,
		branch.SetupGated,
		branchprotection.SetupGated,
		branchprotectionv3.SetupGated,
		defaultbranch.SetupGated,
		deploykey.SetupGated,
		environment.SetupGated,
		environmentdeploymentpolicy.SetupGated,
		githubissue.SetupGated,
		issuelabels.SetupGated,
		pullrequest.SetupGated,
		release.SetupGated,
		repository.SetupGated,
		repositoryautolinkreference.SetupGated,
		repositorycollaborator.SetupGated,
		repositorycollaboratorset.SetupGated,
		repositorycustomproperty.SetupGated,
		repositorydependabotsecurityupdates.SetupGated,
		repositoryfile.SetupGated,
		repositorylabel.SetupGated,
		repositorymilestone.SetupGated,
		repositorypages.SetupGated,
		repositoryruleset.SetupGated,
		repositorytopics.SetupGated,
		repositoryvulnerabilityalerts.SetupGated,
		repositorywebhook.SetupGated,
		emugroupmapping.SetupGated,
		members.SetupGated,
		organizationroleteam.SetupGated,
		team.SetupGated,
		teammembership.SetupGated,
		teamrepository.SetupGated,
		teamsettings.SetupGated,
		teamsyncgroupmapping.SetupGated,
		gpgkey.SetupGated,
		invitationaccepter.SetupGated,
		membership.SetupGated,
		organizationroleuser.SetupGated,
		sshkey.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// SetupWebhookWithManager registers conversion webhooks for all resource kinds in the group.
func SetupWebhookWithManager(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		actionssecret.SetupWebhookWithManager,
		actionsvariable.SetupWebhookWithManager,
		environmentsecret.SetupWebhookWithManager,
		environmentvariable.SetupWebhookWithManager,
		hostedrunner.SetupWebhookWithManager,
		organizationactionssecret.SetupWebhookWithManager,
		organizationactionssecretrepository.SetupWebhookWithManager,
		organizationactionssecretrepositoryset.SetupWebhookWithManager,
		organizationactionsvariable.SetupWebhookWithManager,
		organizationactionsvariablerepository.SetupWebhookWithManager,
		organizationactionsvariablerepositoryset.SetupWebhookWithManager,
		organizationoidcsubjectclaimcustomizationtemplate.SetupWebhookWithManager,
		organizationpermissions.SetupWebhookWithManager,
		organizationworkflowpermissions.SetupWebhookWithManager,
		repositoryaccesslevel.SetupWebhookWithManager,
		repositoryoidcsubjectclaimcustomizationtemplate.SetupWebhookWithManager,
		repositorypermissions.SetupWebhookWithManager,
		repositoryworkflowpermissions.SetupWebhookWithManager,
		runnergroup.SetupWebhookWithManager,
		appinstallationrepositories.SetupWebhookWithManager,
		appinstallationrepositoryassociation.SetupWebhookWithManager,
		codespacessecret.SetupWebhookWithManager,
		organizationcodespacessecret.SetupWebhookWithManager,
		organizationcodespacessecretrepositoryset.SetupWebhookWithManager,
		usercodespacessecret.SetupWebhookWithManager,
		actionsrunnergroup.SetupWebhookWithManager,
		actionsworkflowpermissions.SetupWebhookWithManager,
		ipallowlistentry.SetupWebhookWithManager,
		organization.SetupWebhookWithManager,
		organizationruleset.SetupWebhookWithManager,
		securityanalysissettings.SetupWebhookWithManager,
		organizationblock.SetupWebhookWithManager,
		organizationrepositoryrole.SetupWebhookWithManager,
		organizationrole.SetupWebhookWithManager,
		organizationwebhook.SetupWebhookWithManager,
		providerconfig.SetupWebhookWithManager,
		branch.SetupWebhookWithManager,
		branchprotection.SetupWebhookWithManager,
		branchprotectionv3.SetupWebhookWithManager,
		defaultbranch.SetupWebhookWithManager,
		deploykey.SetupWebhookWithManager,
		environment.SetupWebhookWithManager,
		environmentdeploymentpolicy.SetupWebhookWithManager,
		githubissue.SetupWebhookWithManager,
		issuelabels.SetupWebhookWithManager,
		pullrequest.SetupWebhookWithManager,
		release.SetupWebhookWithManager,
		repository.SetupWebhookWithManager,
		repositoryautolinkreference.SetupWebhookWithManager,
		repositorycollaborator.SetupWebhookWithManager,
		repositorycollaboratorset.SetupWebhookWithManager,
		repositorycustomproperty.SetupWebhookWithManager,
		repositorydependabotsecurityupdates.SetupWebhookWithManager,
		repositoryfile.SetupWebhookWithManager,
		repositorylabel.SetupWebhookWithManager,
		repositorymilestone.SetupWebhookWithManager,
		repositorypages.SetupWebhookWithManager,
		repositoryruleset.SetupWebhookWithManager,
		repositorytopics.SetupWebhookWithManager,
		repositoryvulnerabilityalerts.SetupWebhookWithManager,
		repositorywebhook.SetupWebhookWithManager,
		emugroupmapping.SetupWebhookWithManager,
		members.SetupWebhookWithManager,
		organizationroleteam.SetupWebhookWithManager,
		team.SetupWebhookWithManager,
		teammembership.SetupWebhookWithManager,
		teamrepository.SetupWebhookWithManager,
		teamsettings.SetupWebhookWithManager,
		teamsyncgroupmapping.SetupWebhookWithManager,
		gpgkey.SetupWebhookWithManager,
		invitationaccepter.SetupWebhookWithManager,
		membership.SetupWebhookWithManager,
		organizationroleuser.SetupWebhookWithManager,
		sshkey.SetupWebhookWithManager,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
