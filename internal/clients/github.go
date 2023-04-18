/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"encoding/json"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/upbound/upjet/pkg/terraform"

	"github.com/coopnorge/provider-github/apis/v1beta1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal github credentials as JSON"

	// provider config variables
	keyBaseURL               = "base_url"
	keyOwner                 = "owner"
	keyToken                 = "token"
	keyAppAuth               = "app_auth"
	keyAppAuthID             = "id"
	keyAppAuthInstallationID = "installation_id"
	keyAppAuthPemFile        = "pem_file"
	keyWriteDelayMs          = "write_delay_ms"
	keyReadDelayMs           = "read_delay_ms"
)

type appAuth struct {
	ID             string `json:"id"`
	InstallationID string `json:"installation_id"`
	AuthPemFile    string `json:"pem_file"`
}

type githubConfig struct {
	BaseURL      *string    `json:"base_url,omitempty"`
	Owner        *string    `json:"owner,omitempty"`
	Token        *string    `json:"token,omitempty"`
	AppAuth      *[]appAuth `json:"app_auth,omitempty"`
	WriteDelayMs *int       `json:"write_delay_ms,omitempty"`
	ReadDelayMs  *int       `json:"read_delay_ms,omitempty"`
}

func terraformProviderConfigurationBuilder(creds githubConfig) terraform.ProviderConfiguration {

	cnf := terraform.ProviderConfiguration{}

	if creds.BaseURL != nil {
		cnf[keyBaseURL] = *creds.BaseURL
	}

	if creds.Owner != nil {
		cnf[keyOwner] = *creds.Owner
	}

	if creds.Token != nil {
		cnf[keyToken] = *creds.Token
	}

	if creds.AppAuth != nil {
		aaList := []map[string]any{}

		aa := map[string]any{
			keyAppAuthID:             (*creds.AppAuth)[0].ID,
			keyAppAuthInstallationID: (*creds.AppAuth)[0].InstallationID,
			keyAppAuthPemFile:        (*creds.AppAuth)[0].AuthPemFile,
		}

		aaList = append(aaList, aa)
		cnf[keyAppAuth] = aaList
	}

	if creds.WriteDelayMs != nil {
		cnf[keyWriteDelayMs] = *creds.WriteDelayMs
	}

	if creds.ReadDelayMs != nil {
		cnf[keyReadDelayMs] = *creds.ReadDelayMs
	}

	return cnf

}

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}

		pc := &v1beta1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1beta1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}

		creds := githubConfig{}
		if err := json.Unmarshal(data, &creds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}

		ps.Configuration = terraformProviderConfigurationBuilder(creds)

		return ps, nil

	}
}
