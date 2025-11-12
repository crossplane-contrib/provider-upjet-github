/*
Copyright 2022 Upbound Inc.
*/

package config

import "github.com/crossplane/upjet/v2/pkg/config"

// terraformPluginSDKExternalNameConfigs contains all external name configurations for this
// provider.
var terraformPluginSDKExternalNameConfigs = map[string]config.ExternalName{
	// Can be imported using the following format: {{ runner id }}
	"github_actions_runner_group": config.IdentifierFromProvider,
	// This cannot be imported.
	"github_actions_environment_secret": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ environment }}:{{ variable }}
	"github_actions_environment_variable": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ organization }}
	"github_actions_organization_permissions": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ secret id }}
	"github_actions_organization_secret": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ variable id }}
	"github_actions_organization_variable": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}
	"github_actions_repository_access_level": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}
	"github_actions_repository_permissions": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ secret }}
	"github_actions_secret": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ variable_name }}
	"github_actions_variable": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}:{{ name }}:{{ source branch }}. Using this a id makes
	// the branch field unavailable. This causes the name of the k8s object to be leading and will cause naming conflict.
	"github_branch": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}
	"github_branch_default": config.TemplatedStringAsIdentifier("repository", "{{ .external_name }}"),
	// Imported by using the following format: {{ repository }}:{{ pattern }}
	// We cannot use the external_name = pattern here since pattern can contain non alpha numberic characters
	"github_branch_protection": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}: {{ branch}}
	"github_branch_protection_v3": config.IdentifierFromProvider,
	// Imported by using the following format: {{ group_id }}
	"github_emu_group_mapping": config.IdentifierFromProvider,
	// Imported by using the following format: {{ slug/orgname }} E.g:enterp/some-awesome-org
	"github_enterprise_organization": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository_name }}
	"github_issue_labels": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ organization }}:{{ username }}.
	"github_membership": config.IdentifierFromProvider,
	// Imported by using the following format: {{ ruleset ID }}
	"github_organization_ruleset": config.IdentifierFromProvider,
	// Imported by using the following format: {{name}}
	"github_repository": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}/{{ id }} or {{ key_prefix }}.
	"github_repository_autolink_reference": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ username }}.
	"github_repository_collaborator": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ organization }}:{{ repository }}:{{ property_name }}
	"github_repository_custom_property": config.IdentifierFromProvider,
	// Imported by using the following format: {{ repository }}:{{ (key_id, fetchable from api) }}
	"github_repository_deploy_key": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ environment }}
	"github_repository_environment": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ name }}:{{ environment }}:{{ Id }}
	"github_repository_environment_deployment_policy": config.IdentifierFromProvider,
	// Imported by using the following format: github_repository_file.gitignore {{repository}}/{{file}}:{{branch}}
	// We cannot use file as external name since filenames are not DNSSpec and metadata.name requires this.
	"github_repository_file": config.IdentifierFromProvider,
	// No documentation on how to import
	"github_repository_pull_request": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ ruleset id }}
	"github_repository_ruleset": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}:{{ tag protection id }}
	"github_repository_tag_protection": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ repository }}/{{ id }}.
	"github_repository_webhook": config.IdentifierFromProvider,
	// Imported by using the following format: {{ id / slug }}
	// The id in the state needs to use the numberic id of the team. Cannot make external_name nice
	"github_team": config.IdentifierFromProvider,
	// Can be imported using the following format: {{ team_id }}/{{ team_slug }}.
	"github_team_members": config.IdentifierFromProvider,
	// Imported by using the following format: {{ team_id }}:{{ username }} or {{ team_name }}:{{ username }}
	"github_team_membership": config.IdentifierFromProvider,
	// Imported by using the following format: {{ team_id/slug }}:{{ repository }}
	// The id in the state needs to use the numberic id of the team plus the repository. Cannot make external_name nice
	"github_team_repository": config.IdentifierFromProvider,
	// This is imported using Github Team ID or Team slug.
	"github_team_settings": config.IdentifierFromProvider,
	// Imported by using the following format: {{team_slug}}
	"github_team_sync_group_mapping": config.IdentifierFromProvider,
}

// cliReconciledExternalNameConfigs contains all external name configurations
// belonging to Terraform resources to be reconciled under the CLI-based
// architecture for this provider.
var cliReconciledExternalNameConfigs = map[string]config.ExternalName{}

// resourceConfigurator applies all external name configs
// listed in the table terraformPluginSDKExternalNameConfigs and
// cliReconciledExternalNameConfigs and sets the version
// of those resources to v1beta1. For those resource in
// terraformPluginSDKExternalNameConfigs, it also sets
// config.Resource.UseNoForkClient to `true`.
func resourceConfigurator() config.ResourceOption {
	return func(r *config.Resource) {
		// if configured both for the no-fork and CLI based architectures,
		// no-fork configuration prevails
		e, configured := terraformPluginSDKExternalNameConfigs[r.Name]
		if !configured {
			e, configured = cliReconciledExternalNameConfigs[r.Name]
		}
		if !configured {
			return
		}
		r.Version = "v1alpha1"
		r.ExternalName = e
	}
}
