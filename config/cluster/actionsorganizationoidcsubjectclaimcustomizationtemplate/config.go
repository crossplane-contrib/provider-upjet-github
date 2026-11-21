package actionsorganizationoidcsubjectclaimcustomizationtemplate

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_oidc_subject_claim_customization_template resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_oidc_subject_claim_customization_template", func(r *config.Resource) {
		r.Kind = "OrganizationOIDCSubjectClaimCustomizationTemplate"
		r.ShortGroup = "actions"
	})
}
