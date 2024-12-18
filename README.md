# Provider GitHub

`provider-upjet-github` is a [Crossplane](https://crossplane.io/) provider that
is built using [Upjet](https://github.com/upbound/upjet) code
generation tools and exposes XRM-conformant managed resources for the
GitHub API.

## Getting Started

Install the provider by using the following command after changing the image tag
to the [latest release](https://marketplace.upbound.io/providers/crossplane-contrib/provider-upjet-github):
```
up ctp provider install crossplane-contrib/provider-upjet-github:v0.1.0
```

Alternatively, you can use declarative installation:
```
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-upjet-github
spec:
  package: crossplane-contrib/provider-upjet-github:v0.1.0
EOF
```
You can see the API reference [here](https://marketplace.upbound.io/providers/crossplane-contrib/provider-upjet-github).

### Adding provider config

Add this to configure the provider. Reference on how to configure this
can be found at the terraform provider documentation
https://registry.terraform.io/providers/integrations/github/latest/docs

#### Provider config example with personal access token

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: provider-secret
  namespace: upbound-system
type: Opaque
stringData:
  credentials: "{\"token\":\"${GH_TOKEN}\",\"owner\":\"${GH_OWNER}\"}"

---
apiVersion: github.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: provider-secret
      namespace: upbound-system
      key: credentials

```

#### Provider config example with Github application based authentication

Note that the PEM certificate needs to be wrapped in a non-multiline string, with the characters "\n"
as newline. See Terraform provider doc for more information.
```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: provider-secret
  namespace: upbound-system
type: Opaque
stringData:
  credentials: "{\"app_auth\": [{\"id\": \"${APP_ID}\", \"installation_id\": \"${APP_INSTALLATION_ID}\", \"pem_file\": \"${APP_PEM_FILE}\"}] ,\"owner\":\"${GH_OWNER}\"}"

---
apiVersion: github.upbound.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: provider-secret
      namespace: upbound-system
      key: credentials

```

## Supported resources

| Kind | Group | Terraform Resource Name | Notes  |
| ---- | ----- | ----------------------- | ------ |
| `Repository` | `repo` | `github_repository` |  |
| `Branch` | `repo` |  `github_branch`      |  |
| `DefaultBranch` | `repo` | `github_branch_default` | name change |
| `BranchProtection` | `repo` | `github_branch_protection` | |
| `RepositoryFile` | `repo` | `github_repository_file` | |
| `RepositoryAutolinkReference` | `repo` | `github_repository_autolink_reference` | |
| `RepositoryCollaborator` | `repo` | `github_repository_collaborator` | |
| `RepositoryWebhook` | `repo` | `github_repository_webhook` | |
| `RepositoryRuleset` | `repo` | `github_repository_ruleset` | |
| `PullRequest` | `repo` | `github_repository_pull_request` | |
| `DeployKey` | `repo` | `github_repository_deploy_key` | |
| `Team` | `team` | `github_team` | |
| `TeamMembership` | `team` | `github_team_membership` | |
| `TeamRepository` | `team` | `github_team_repository` | |
| `TeamSettings` | `team` | `github_team_settings` | |
| `TeamSyncGroupMapping` | `team` | `github_team_sync_group_mapping` | |
| `EmuTeamMapping` | `team` | `github_emu_group_mapping` | |
| `ActionsSecrets` | `actions` | `github_actions_secret` | |
| `OrganizationRuleset` | `organization` | `github_organization_ruleset` | |
| `Membership` | `user` | `github_membership` | Works only with - GitHub App user access tokens - GitHub App installation access tokens - Fine-grained personal access tokens using a token with at least `members:read` |
| `Organization` | `enterprise` | `github_enterprise_organization` | Works only with PAT based authentication using a token with at least `admin:enterprise` scope.<br/>The following *GraphQL* query can be used to obtain the required `enterprise_id`:<br/> `gh api graphql -f query='query ($slug: String!) { enterprise(slug: $slug) { id } }' -F slug='<slug>' --jq '.data.enterprise.id'` |

## Adding resources

* Find the resource to add here: https://registry.terraform.io/providers/integrations/github/latest/docs
* 1 resource per PR prefered
* Write a test case

Check this reference PR: https://github.com/crossplane-contrib/provider-upjet-github/pull/4

An example diff for human generated files

```diff
diff --git a/README.md b/README.md
index 06704c1..7adefad 100644
--- a/README.md
+++ b/README.md
@@ -34,6 +34,7 @@ You can see the API reference [here](https://doc.crds.dev/github.com/crossplane-contrib/p
 | `Branch` | `repo` |  `github_branch`      |  |
 | `DefaultBranch` | `repo` | `github_branch_default` | name change |
 | `BranchProtection` | `repo` | `github_branch_protection` | |
+| `RepositoryFile` | `repo` | `github_repository_file` | |
 | `Team` | `team` | `github_team` | |
 | `TeamRepository` | `team` | `github_team_repository` | |

diff --git a/config/external_name.go b/config/external_name.go
index 505fa1c..50440d4 100644
--- a/config/external_name.go
+++ b/config/external_name.go
@@ -18,6 +18,9 @@ var ExternalNameConfigs = map[string]config.ExternalName{
 	// Imported by using the following format: {{ repository }}:{{ pattern }}
 	// We cannot use the external_name = pattern here since pattern can contain non alpha numberic characters
 	"github_branch_protection": config.IdentifierFromProvider,
+  // Imported by using the following format: github_repository_file.gitignore {{repository}}/{{file}}:{{branch}}
+  // We cannot use file as external name since filenames are not DNSSpec and metadata.name requires this.
+  "github_repository_file": config.IdentifierFromProvider,
 	// Imported by using the following format: {{ id / slug }}
 	// The id in the state needs to use the numberic id of the team. Cannot make external_name nice
 	"github_team": config.IdentifierFromProvider,
diff --git a/config/provider.go b/config/provider.go
index e2d81bf..093bdf8 100644
--- a/config/provider.go
+++ b/config/provider.go
@@ -12,6 +12,7 @@ import (
 	"github.com/crossplane-contrib/provider-upjet-github/config/branchprotection"
 	"github.com/crossplane-contrib/provider-upjet-github/config/defaultbranch"
 	"github.com/crossplane-contrib/provider-upjet-github/config/repository"
+	"github.com/crossplane-contrib/provider-upjet-github/config/repositoryfile"
 	"github.com/crossplane-contrib/provider-upjet-github/config/team"
 	"github.com/crossplane-contrib/provider-upjet-github/config/teamrepository"
 	ujconfig "github.com/upbound/upjet/pkg/config"
@@ -40,6 +41,7 @@ func GetProvider() *ujconfig.Provider {
 		// add custom config functions
 		repository.Configure,
 		branch.Configure,
+    repositoryfile.Configure,
 		team.Configure,
 		teamrepository.Configure,
 		defaultbranch.Configure,
diff --git a/config/repositoryfile/config.go b/config/repositoryfile/config.go
new file mode 100644
index 0000000..5684451
--- /dev/null
+++ b/config/repositoryfile/config.go
@@ -0,0 +1,20 @@
+package repositoryfile
+
+import "github.com/upbound/upjet/pkg/config"
+
+// Configure github_repository_file resource.
+func Configure(p *config.Provider) {
+	p.AddResourceConfigurator("github_repository_file", func(r *config.Resource) {
+		// We need to override the default group that upjet generated for
+		// this resource, which would be "github"
+		r.Kind = "RepositoryFile"
+		r.ShortGroup = "repo"
+
+    r.References["repository"] = config.Reference{
+			Type: "Repository",
+		}
+    r.References["branch"] = config.Reference{
+			Type: "Branch",
+		}
+	})
+}
diff --git a/tests/cases/repo-branch-file/repository.yaml b/tests/cases/repo-branch-file/repository.yaml
new file mode 100644
index 0000000..b1c6cf4
--- /dev/null
+++ b/tests/cases/repo-branch-file/repository.yaml
@@ -0,0 +1,44 @@
+---
+apiVersion: repo.github.upbound.io/v1alpha1
+kind: Repository
+metadata:
+  name: github-crossplane-file-test
+spec:
+  forProvider:
+    visibility: public
+    autoInit: true
+    gitignoreTemplate: Terraform
+  providerConfigRef:
+    name: default
+---
+apiVersion: repo.github.upbound.io/v1alpha1
+kind: Branch
+metadata:
+  name: add-file-branch
+spec:
+  forProvider:
+    repositoryRef:
+      name: github-crossplane-file-test
+  providerConfigRef:
+    name: default
+---
+apiVersion: repo.github.upbound.io/v1alpha1
+kind: RepositoryFile
+metadata:
+  name: sample-file-dot-txt
+spec:
+  forProvider:
+    file: sample-file.txt
+    content: |
+      I am an crossplane 
+      provider. This should be nice.
+    commitMessage: "managed by crossplane provider"
+    commitAuthor: "Crossplane Github Provider"
+    commitEmail: "github-provider@crossplane.com"
+    repositoryRef:
+      name: github-crossplane-file-test
+    branchRef:
+      name: add-file-branch
+  providerConfigRef:
+    name: default
+
```



