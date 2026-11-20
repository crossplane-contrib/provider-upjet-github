package enterpriseactionsworkflowpermissions

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_enterprise_actions_workflow_permissions resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_enterprise_actions_workflow_permissions", func(r *config.Resource) {
		r.Kind = "ActionsWorkflowPermissions"
		r.ShortGroup = "enterprise"
	})
}
