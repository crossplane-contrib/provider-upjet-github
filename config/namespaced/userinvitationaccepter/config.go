package userinvitationaccepter

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_user_invitation_accepter resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_user_invitation_accepter", func(r *config.Resource) {
		r.Kind = "InvitationAccepter"
		r.ShortGroup = "user"

		r.References["invitation_id"] = config.Reference{
			TerraformName: "github_repository_collaborator",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("invitation_id",true)`,
		}
	})
}
