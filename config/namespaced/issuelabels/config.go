package issuelabels

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_issue_label resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_issue_labels", func(r *config.Resource) {
		r.Kind = "IssueLabels"
		r.ShortGroup = "repo"

		r.References["repository "] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
