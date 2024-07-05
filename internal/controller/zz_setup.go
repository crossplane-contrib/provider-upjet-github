/*
Copyright 2022 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	actionssecret "github.com/coopnorge/provider-github/internal/controller/actions/actionssecret"
	actionsvariable "github.com/coopnorge/provider-github/internal/controller/actions/actionsvariable"
	organization "github.com/coopnorge/provider-github/internal/controller/enterprise/organization"
	organizationruleset "github.com/coopnorge/provider-github/internal/controller/enterprise/organizationruleset"
	providerconfig "github.com/coopnorge/provider-github/internal/controller/providerconfig"
	branch "github.com/coopnorge/provider-github/internal/controller/repo/branch"
	branchprotection "github.com/coopnorge/provider-github/internal/controller/repo/branchprotection"
	defaultbranch "github.com/coopnorge/provider-github/internal/controller/repo/defaultbranch"
	deploykey "github.com/coopnorge/provider-github/internal/controller/repo/deploykey"
	pullrequest "github.com/coopnorge/provider-github/internal/controller/repo/pullrequest"
	repository "github.com/coopnorge/provider-github/internal/controller/repo/repository"
	repositoryfile "github.com/coopnorge/provider-github/internal/controller/repo/repositoryfile"
	repositorywebhook "github.com/coopnorge/provider-github/internal/controller/repo/repositorywebhook"
	emugroupmapping "github.com/coopnorge/provider-github/internal/controller/team/emugroupmapping"
	team "github.com/coopnorge/provider-github/internal/controller/team/team"
	teammembership "github.com/coopnorge/provider-github/internal/controller/team/teammembership"
	teamrepository "github.com/coopnorge/provider-github/internal/controller/team/teamrepository"
	teamsettings "github.com/coopnorge/provider-github/internal/controller/team/teamsettings"
	teamsyncgroupmapping "github.com/coopnorge/provider-github/internal/controller/team/teamsyncgroupmapping"
	membership "github.com/coopnorge/provider-github/internal/controller/user/membership"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		actionssecret.Setup,
		actionsvariable.Setup,
		organization.Setup,
		organizationruleset.Setup,
		providerconfig.Setup,
		branch.Setup,
		branchprotection.Setup,
		defaultbranch.Setup,
		deploykey.Setup,
		pullrequest.Setup,
		repository.Setup,
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
