package enterpriseipallowlistentry

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_enterprise_ip_allow_list_entry resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_enterprise_ip_allow_list_entry", func(r *config.Resource) {
		r.Kind = "IPAllowListEntry"
		r.ShortGroup = "enterprise"
	})
}
