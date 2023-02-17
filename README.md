# Provider GitHub

`provider-github` is a [Crossplane](https://crossplane.io/) provider that
is built using [Upjet](https://github.com/upbound/upjet) code
generation tools and exposes XRM-conformant managed resources for the
GitHub API.

## Getting Started

Install the provider by using the following command after changing the image tag
to the [latest release](https://marketplace.upbound.io/providers/coopnorge/provider-github):
```
up ctp provider install coopnorge/provider-github:v0.1.0
```

Alternatively, you can use declarative installation:
```
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-github
spec:
  package: coopnorge/provider-github:v0.1.0
EOF
```
You can see the API reference [here](https://doc.crds.dev/github.com/coopnorge/provider-github).

## Supported resources

| Kind | Group | Terraform Resource Name | Notes  |
| ---- | ----- | ----------------------- | ------ |
| `Repository` | `repo` | `github_repository` |  |
| `Branch` | `repo` |  `github_branch`      |  |
| `DefaultBranch` | `repo` | `github_branch_default` | name change |
| `BranchProtection` | `repo` | `github_branch_protection` | |
| `RepositoryFile` | `repo` | `github_repository_file` | |
| `PullRequest` | `repo` | `github_repository_pull_request` | |
| `Team` | `team` | `github_team` | | 
| `TeamRepository` | `team` | `github_team_repository` | |
| ActionsSecrets | actions | github_actions_secret | | 


## Adding resources

* Find the resource to add here: https://registry.terraform.io/providers/integrations/github/latest/docs
* 1 resource per PR prefered
* Write a test case 

Check this reference PR: https://github.com/coopnorge/provider-github/pull/4 

An example diff for human generated files

```diff
diff --git a/README.md b/README.md
index 06704c1..7adefad 100644
--- a/README.md
+++ b/README.md
@@ -34,6 +34,7 @@ You can see the API reference [here](https://doc.crds.dev/github.com/coopnorge/p
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
 	"github.com/coopnorge/provider-github/config/branchprotection"
 	"github.com/coopnorge/provider-github/config/defaultbranch"
 	"github.com/coopnorge/provider-github/config/repository"
+	"github.com/coopnorge/provider-github/config/repositoryfile"
 	"github.com/coopnorge/provider-github/config/team"
 	"github.com/coopnorge/provider-github/config/teamrepository"
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
+// Configure github_branch resource.
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



