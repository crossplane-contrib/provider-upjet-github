package actionsorganizationworkflowpermissions

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_workflow_permissions resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_workflow_permissions", func(r *config.Resource) {
		r.Kind = "OrganizationWorkflowPermissions"
		r.ShortGroup = "actions"
	})
}
