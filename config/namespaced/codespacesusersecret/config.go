package codespacesusersecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_codespaces_user_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_codespaces_user_secret", func(r *config.Resource) {
		r.Kind = "UserCodespacesSecret"
		r.ShortGroup = "codespaces"
	})
}
