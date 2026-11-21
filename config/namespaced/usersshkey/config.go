package usersshkey

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_user_ssh_key resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_user_ssh_key", func(r *config.Resource) {
		r.Kind = "SSHKey"
		r.ShortGroup = "user"
	})
}
