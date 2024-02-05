// SPDX-FileCopyrightText: 2023 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/upjet/pkg/controller"

	actionssecret "github.com/coopnorge/provider-github/internal/controller/actions/actionssecret"
	organization "github.com/coopnorge/provider-github/internal/controller/organization/organization"
	providerconfig "github.com/coopnorge/provider-github/internal/controller/providerconfig"
	branch "github.com/coopnorge/provider-github/internal/controller/repo/branch"
	branchprotection "github.com/coopnorge/provider-github/internal/controller/repo/branchprotection"
	defaultbranch "github.com/coopnorge/provider-github/internal/controller/repo/defaultbranch"
	deploykey "github.com/coopnorge/provider-github/internal/controller/repo/deploykey"
	pullrequest "github.com/coopnorge/provider-github/internal/controller/repo/pullrequest"
	repository "github.com/coopnorge/provider-github/internal/controller/repo/repository"
	repositoryfile "github.com/coopnorge/provider-github/internal/controller/repo/repositoryfile"
	team "github.com/coopnorge/provider-github/internal/controller/team/team"
	teamrepository "github.com/coopnorge/provider-github/internal/controller/team/teamrepository"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		actionssecret.Setup,
		organization.Setup,
		providerconfig.Setup,
		branch.Setup,
		branchprotection.Setup,
		defaultbranch.Setup,
		deploykey.Setup,
		pullrequest.Setup,
		repository.Setup,
		repositoryfile.Setup,
		team.Setup,
		teamrepository.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
