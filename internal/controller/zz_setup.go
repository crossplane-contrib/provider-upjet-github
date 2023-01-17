/*
Copyright 2021 Upbound Inc.
*/

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/upbound/upjet/pkg/controller"

	protection "github.com/coopnorge/provider-github/internal/controller/branch/protection"
	team "github.com/coopnorge/provider-github/internal/controller/github/team"
	providerconfig "github.com/coopnorge/provider-github/internal/controller/providerconfig"
	branch "github.com/coopnorge/provider-github/internal/controller/repo/branch"
	repository "github.com/coopnorge/provider-github/internal/controller/repo/repository"
	repositoryteam "github.com/coopnorge/provider-github/internal/controller/team/repository"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		protection.Setup,
		team.Setup,
		providerconfig.Setup,
		branch.Setup,
		repository.Setup,
		repositoryteam.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
