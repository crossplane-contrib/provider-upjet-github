package repository

import (
	"github.com/crossplane/upjet/v2/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Configure github_repository resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository", func(r *config.Resource) {
		r.ShortGroup = "repo"

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"private", "default_branch"},
		}

		r.TerraformCustomDiff = func(diff *terraform.InstanceDiff, _ *terraform.InstanceState, _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
			if diff == nil || diff.Destroy {
				return diff, nil
			}
			if ppDiff, ok := diff.Attributes["security_and_analysis.#"]; ok && ppDiff.Old == "" && ppDiff.New == "" {
				delete(diff.Attributes, "security_and_analysis.#")
			}
			if ppDiff, ok := diff.Attributes["topics.#"]; ok && ppDiff.Old == "" && ppDiff.New == "" {
				delete(diff.Attributes, "topics.#")
			}

			return diff, nil
		}
	})
}
