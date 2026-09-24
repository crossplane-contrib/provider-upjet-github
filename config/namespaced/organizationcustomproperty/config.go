package organizationcustomproperty

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_custom_properties resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_custom_properties", func(r *config.Resource) {
		// We need to override the default group/kind that upjet generated
		// for this resource, which would be group "github" and kind
		// "OrganizationCustomProperties". The Terraform resource manages a
		// single custom property despite its plural name.
		r.Kind = "OrganizationCustomProperty"
		r.ShortGroup = "organization"
	})
}
