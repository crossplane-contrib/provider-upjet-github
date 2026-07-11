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

		r.TerraformCustomDiff = repositoryCustomDiff
	})
}

// repositoryCustomDiff sanitizes the computed Terraform diff for
// github_repository resources.
func repositoryCustomDiff(diff *terraform.InstanceDiff, _ *terraform.InstanceState, _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
	if diff == nil || diff.Destroy {
		return diff, nil
	}
	// Drop spurious no-op diffs on computed/optional collections.
	dropAttrDiff(diff, "security_and_analysis.#", "", "")
	dropAttrDiff(diff, "topics.#", "", "")
	// GitHub does not support unarchiving a repository through this resource, so
	// an archived repository whose desired state omits archived=true produces a
	// perpetual archived: true->false diff. Beyond being a no-op update loop,
	// this diff leaks into the destroy path: the Terraform delete reads the
	// diff's archived=false and re-PATCHes the already-archived (read-only)
	// repository, which returns HTTP 403 and wedges deletion forever. Drop it.
	dropAttrDiff(diff, "archived", "true", "false")

	return diff, nil
}

// dropAttrDiff removes key from the diff when its computed change matches the
// given oldVal->newVal transition, suppressing a spurious diff.
func dropAttrDiff(diff *terraform.InstanceDiff, key, oldVal, newVal string) {
	if d, ok := diff.Attributes[key]; ok && d.Old == oldVal && d.New == newVal {
		delete(diff.Attributes, key)
	}
}
