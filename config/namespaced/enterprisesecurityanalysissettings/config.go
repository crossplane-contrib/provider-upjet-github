package enterprisesecurityanalysissettings

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_enterprise_security_analysis_settings resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_enterprise_security_analysis_settings", func(r *config.Resource) {
		r.Kind = "SecurityAnalysisSettings"
		r.ShortGroup = "enterprise"
	})
}
