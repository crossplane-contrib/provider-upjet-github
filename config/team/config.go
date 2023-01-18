package team

import "github.com/upbound/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_team", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "Team"
		r.ShortGroup = "team"

		// This resource need the repository in which branch would be created
		// as an input. And by defining it as a reference to Repository
		// object, we can build cross resource referencing. See
		// repositoryRef in the example in the Testing section below.
		r.References["repository"] = config.Reference{
			Type: "github.com/coopnorge/provider-github/apis/repo/v1alpha1.Repository",
		}
	})
}
