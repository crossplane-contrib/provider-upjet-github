/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	actionssecret "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/actionssecret"
	actionsvariable "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/actionsvariable"
	runnergroup "github.com/crossplane-contrib/provider-upjet-github/internal/controller/actions/runnergroup"
	organization "github.com/crossplane-contrib/provider-upjet-github/internal/controller/enterprise/organization"
	organizationruleset "github.com/crossplane-contrib/provider-upjet-github/internal/controller/enterprise/organizationruleset"
	providerconfig "github.com/crossplane-contrib/provider-upjet-github/internal/controller/providerconfig"
	branch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/branch"
	branchprotection "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/branchprotection"
	defaultbranch "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/defaultbranch"
	deploykey "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/deploykey"
	environment "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/environment"
	pullrequest "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/pullrequest"
	repository "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repository"
	repositoryautolinkreference "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositoryautolinkreference"
	repositorycollaborator "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositorycollaborator"
	repositoryfile "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositoryfile"
	repositorywebhook "github.com/crossplane-contrib/provider-upjet-github/internal/controller/repo/repositorywebhook"
	emugroupmapping "github.com/crossplane-contrib/provider-upjet-github/internal/controller/team/emugroupmapping"
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
		runnergroup.Setup,
		organization.Setup,
		organizationruleset.Setup,
		providerconfig.Setup,
		branch.Setup,
		branchprotection.Setup,
		defaultbranch.Setup,
		deploykey.Setup,
		environment.Setup,
		pullrequest.Setup,
		repository.Setup,
		repositoryautolinkreference.Setup,
		repositorycollaborator.Setup,
		repositoryfile.Setup,
		repositorywebhook.Setup,
		emugroupmapping.Setup,
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
