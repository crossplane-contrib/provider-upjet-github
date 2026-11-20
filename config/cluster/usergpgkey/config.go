package usergpgkey

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_user_gpg_key resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_user_gpg_key", func(r *config.Resource) {
		r.Kind = "GPGKey"
		r.ShortGroup = "user"
	})
}
