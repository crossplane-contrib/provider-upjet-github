/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/v2/pkg/controller"

	actionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/actionssecret"
	actionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/actionsvariable"
	environmentsecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/environmentsecret"
	environmentvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/environmentvariable"
	organizationactionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/organizationactionssecret"
	organizationactionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/organizationactionsvariable"
	organizationpermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/organizationpermissions"
	repositoryaccesslevel "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/repositoryaccesslevel"
	repositorypermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/repositorypermissions"
	runnergroup "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/actions/runnergroup"
	organization "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/enterprise/organization"
	organizationruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/enterprise/organizationruleset"
	providerconfig "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/providerconfig"
	branch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/branch"
	branchprotection "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/branchprotection"
	branchprotectionv3 "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/branchprotectionv3"
	defaultbranch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/defaultbranch"
	deploykey "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/deploykey"
	environment "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/environment"
	environmentdeploymentpolicy "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/environmentdeploymentpolicy"
	issuelabels "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/issuelabels"
	pullrequest "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/pullrequest"
	repository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repository"
	repositoryautolinkreference "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repositoryautolinkreference"
	repositorycollaborator "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repositorycollaborator"
	repositorycustomproperty "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repositorycustomproperty"
	repositoryfile "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repositoryfile"
	repositoryruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repositoryruleset"
	repositorywebhook "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/repo/repositorywebhook"
	emugroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/emugroupmapping"
	members "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/members"
	team "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/team"
	teammembership "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/teammembership"
	teamrepository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/teamrepository"
	teamsettings "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/teamsettings"
	teamsyncgroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/team/teamsyncgroupmapping"
	membership "github.com/crossplane-contrib/provider-upjet-github/internal/controller/namespaced/user/membership"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		actionssecret.Setup,
		actionsvariable.Setup,
		environmentsecret.Setup,
		environmentvariable.Setup,
		organizationactionssecret.Setup,
		organizationactionsvariable.Setup,
		organizationpermissions.Setup,
		repositoryaccesslevel.Setup,
		repositorypermissions.Setup,
		runnergroup.Setup,
		organization.Setup,
		organizationruleset.Setup,
		providerconfig.Setup,
		branch.Setup,
		branchprotection.Setup,
		branchprotectionv3.Setup,
		defaultbranch.Setup,
		deploykey.Setup,
		environment.Setup,
		environmentdeploymentpolicy.Setup,
		issuelabels.Setup,
		pullrequest.Setup,
		repository.Setup,
		repositoryautolinkreference.Setup,
		repositorycollaborator.Setup,
		repositorycustomproperty.Setup,
		repositoryfile.Setup,
		repositoryruleset.Setup,
		repositorywebhook.Setup,
		emugroupmapping.Setup,
		members.Setup,
		team.Setup,
		teammembership.Setup,
		teamrepository.Setup,
		teamsettings.Setup,
		teamsyncgroupmapping.Setup,
		membership.Setup,
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
		organizationactionssecret.SetupGated,
		organizationactionsvariable.SetupGated,
		organizationpermissions.SetupGated,
		repositoryaccesslevel.SetupGated,
		repositorypermissions.SetupGated,
		runnergroup.SetupGated,
		organization.SetupGated,
		organizationruleset.SetupGated,
		providerconfig.SetupGated,
		branch.SetupGated,
		branchprotection.SetupGated,
		branchprotectionv3.SetupGated,
		defaultbranch.SetupGated,
		deploykey.SetupGated,
		environment.SetupGated,
		environmentdeploymentpolicy.SetupGated,
		issuelabels.SetupGated,
		pullrequest.SetupGated,
		repository.SetupGated,
		repositoryautolinkreference.SetupGated,
		repositorycollaborator.SetupGated,
		repositorycustomproperty.SetupGated,
		repositoryfile.SetupGated,
		repositoryruleset.SetupGated,
		repositorywebhook.SetupGated,
		emugroupmapping.SetupGated,
		members.SetupGated,
		team.SetupGated,
		teammembership.SetupGated,
		teamrepository.SetupGated,
		teamsettings.SetupGated,
		teamsyncgroupmapping.SetupGated,
		membership.SetupGated,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
