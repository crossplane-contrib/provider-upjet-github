package organizationwebhook

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure github_organization_webhook resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_webhook", func(r *config.Resource) {
		r.Kind = "OrganizationWebhook"
		r.ShortGroup = "enterprise"
	})
}
