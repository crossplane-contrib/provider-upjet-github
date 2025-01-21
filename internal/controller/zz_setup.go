/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	actionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/actionssecret"
	actionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/actionsvariable"
	environmentsecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/environmentsecret"
	environmentvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/environmentvariable"
	organizationactionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/organizationactionssecret"
	organizationactionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/organizationactionsvariable"
	organizationpermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/organizationpermissions"
	repositoryaccesslevel "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/repositoryaccesslevel"
	repositorypermissions "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/repositorypermissions"
	runnergroup "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/runnergroup"
	organization "github.com/crossplane-contrib/provider-upjet-github/internal/controller/enterprise/organization"
	organizationruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/enterprise/organizationruleset"
	providerconfig "github.com/crossplane-contrib/provider-upjet-github/internal/controller/providerconfig"
	branch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/branch"
	branchprotection "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/branchprotection"
	branchprotectionv3 "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/branchprotectionv3"
	defaultbranch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/defaultbranch"
	deploykey "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/deploykey"
	environment "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/environment"
	environmentdeploymentpolicy "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/environmentdeploymentpolicy"
	issuelabels "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/issuelabels"
	pullrequest "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/pullrequest"
	repository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repository"
	repositoryautolinkreference "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositoryautolinkreference"
	repositorycollaborator "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositorycollaborator"
	repositoryfile "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositoryfile"
	repositoryruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositoryruleset"
	repositorywebhook "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositorywebhook"
	tagprotection "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/tagprotection"
	emugroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/emugroupmapping"
	members "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/members"
	team "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/team"
	teammembership "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/teammembership"
	teamrepository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/teamrepository"
	teamsettings "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/teamsettings"
	teamsyncgroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/teamsyncgroupmapping"
	membership "github.com/crossplane-contrib/provider-upjet-github/internal/controller/user/membership"
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
		repositoryfile.Setup,
		repositoryruleset.Setup,
		repositorywebhook.Setup,
		tagprotection.Setup,
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
