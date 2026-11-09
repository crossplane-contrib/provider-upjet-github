package repository

import "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

// PreserveArchivedRepositoryState removes the diff that would unarchive a
// repository during deletion when archive_on_destroy is enabled.
func PreserveArchivedRepositoryState(diff *terraform.InstanceDiff, state *terraform.InstanceState, _ *terraform.ResourceConfig) (*terraform.InstanceDiff, error) {
	if diff == nil || state == nil || diff.Destroy {
		return diff, nil
	}
	if state.Attributes["archive_on_destroy"] != "true" || state.Attributes["archived"] != "true" {
		return diff, nil
	}

	if archived, ok := diff.Attributes["archived"]; ok && archived.New == "false" {
		delete(diff.Attributes, "archived")
	}
	return diff, nil
}
