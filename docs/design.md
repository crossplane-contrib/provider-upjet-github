## Layout and naming of resources

### Terraform mapping
```mermaid
graph LR
    github_repository --> github_team_repository
    github_repository --> branch
    github_repository --> github_action_secret
    subgraph repository
        github_repository --> github_repository_webhook
        github_repository --> github_repository_deploy_key
        subgraph branch
            github_branch --> github_branch_default
            github_branch_protection
        end
        subgraph actions
            github_action_secret
        end
    end
    subgraph teams
        github_team --> github_team_members
        github_team --> github_team_settings
        github_team --> github_team_repository
    end
```

### Proposed mapping in github provider

```mermaid
graph LR
    Team.team.provider-github --> TeamAccess.repo.provider-github
    subgraph repo
        Repo.repo.provider-github --> Webhook.repo.provider-github
        Repo.repo.provider-github --> DeployKey.repo.provider-github
        Repo.repo.provider-github --> Branch.repo.provider-github
        Repo.repo.provider-github --> DefaultBranch.repo.provider-github
        Repo.repo.provider-github --> BranchProtection.repo.provider-gitub
        Repo.repo.provider-github --> ActionSecret.repo.provider-github
        Repo.repo.provider-github --> TeamAccess.repo.provider-github
    end
    subgraph team
        Team.team.provider-github --> TeamSettings.team.provider-github
        Team.team.provider-github --> TeamMembers.team.provider-github
    end
```