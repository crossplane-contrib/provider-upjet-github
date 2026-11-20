package workflowrepositorypermissions

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_workflow_repository_permissions resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_workflow_repository_permissions", func(r *config.Resource) {
		r.Kind = "RepositoryWorkflowPermissions"
		r.ShortGroup = "actions"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
