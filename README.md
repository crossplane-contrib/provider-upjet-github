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
| `Team` | `team` | `github_team` | | 
| `TeamRepository` | `team` | `github_team_repository` | |

