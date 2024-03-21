/*
Copyright 2022 Upbound Inc.
*/

package config

import "github.com/crossplane/upjet/pkg/config"

// ExternalNameConfigs contains all external name configurations for this
// provider.
var ExternalNameConfigs = map[string]config.ExternalName{
	// Imported by using the following format: {{name}}
	"github_repository": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}:{{ name }}:{{ source branch }}. Using this a id makes
	// the branch field unavailable. This causes the name of the k8s object to be leading and will cause naming conflict.
	"github_branch": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}
	"github_branch_default": config.TemplatedStringAsIdentifier("repository", "{{ .external_name }}"),
	// Imported by using the following format: {{ repository }}:{{ (key_id, fetchable from api) }}
	"github_repository_deploy_key": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}:{{ pattern }}
	// We cannot use the external_name = pattern here since pattern can contain non alpha numberic characters
	"github_branch_protection": config.IdentifierFromProvider,
	// No documentation on how to import
	"github_repository_pull_request": config.IdentifierFromProvider,
	// Imported by using the following format: github_repository_file.gitignore {{repository}}/{{file}}:{{branch}}
	// We cannot use file as external name since filenames are not DNSSpec and metadata.name requires this.
	"github_repository_file": config.IdentifierFromProvider,
	// Imported by using the following format: {{ id / slug }}
	// The id in the state needs to use the numberic id of the team. Cannot make external_name nice
	"github_team": config.IdentifierFromProvider,
	// Imported by using the following format: {{ team_id/slug }}:{{ repository }}
	// The id in the state needs to use the numberic id of the team plus the repository. Cannot make external_name nice
	"github_team_repository": config.IdentifierFromProvider,
	// This cannot be imported.
	"github_emu_group_mapping": config.IdentifierFromProvider,
	// Imported by using the following format: {{ team_id }}:{{ username }} or {{ team_name }}:{{ username }}
	"github_team_membership": config.IdentifierFromProvider,
	// This is imported using Github Team ID or Team slug.
	"github_team_settings": config.IdentifierFromProvider,
	// This cannot be imported.
	"github_team_sync_group_mapping": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ organization }}:{{ username }}.
	"github_membership": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}/{{ id }}.
	"github_repository_webhook": config.IdentifierFromProvider,
	// This cannot be imported.
	"github_actions_secret":          config.IdentifierFromProvider,
	"github_actions_variable":        config.IdentifierFromProvider,
	"github_enterprise_organization": config.IdentifierFromProvider,
}

// ExternalNameConfigurations applies all external name configs listed in the
// table ExternalNameConfigs and sets the version of those resources to v1beta1
// assuming they will be tested.
func ExternalNameConfigurations() config.ResourceOption {
	return func(r *config.Resource) {
		if e, ok := ExternalNameConfigs[r.Name]; ok {
			r.ExternalName = e
		}
	}
}

// ExternalNameConfigured returns the list of all resources whose external name
// is configured manually.
func ExternalNameConfigured() []string {
	l := make([]string, len(ExternalNameConfigs))
	i := 0
	for name := range ExternalNameConfigs {
		// $ is added to match the exact string since the format is regex.
		l[i] = name + "$"
		i++
	}
	return l
}
