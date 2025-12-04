/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/upjet/pkg/terraform"

	"github.com/crossplane-contrib/provider-upjet-github/apis/v1beta1"
)

const (
	// error messages
	errNoProviderConfig              = "no providerConfigRef provided"
	errGetProviderConfig             = "cannot get referenced ProviderConfig"
	errTrackUsage                    = "cannot track ProviderConfig usage"
	errExtractCredentials            = "cannot extract credentials"
	errUnmarshalCredentials          = "cannot unmarshal github credentials as JSON"
	errProviderConfigurationBuilder  = "cannot build configuration for terraform provider block"
	errTerraformProviderMissingOwner = "github provider app_auth needs owner key to be set"
	errLockError                     = "unable to lock for github credential refresh"

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
	keyRetryDelayMs          = "retry_delay_ms"
	keyMaxRetries            = "max_retries"
	keyRetryableErrors       = "retryable_errors"
)

type appAuth struct {
	ID             string `json:"id"`
	InstallationID string `json:"installation_id"`
	AuthPemFile    string `json:"pem_file"`
}

type githubConfig struct {
	BaseURL         *string    `json:"base_url,omitempty"`
	Owner           *string    `json:"owner,omitempty"`
	Token           *string    `json:"token,omitempty"`
	AppAuth         *[]appAuth `json:"app_auth,omitempty"`
	WriteDelayMs    *int       `json:"write_delay_ms,omitempty"`
	ReadDelayMs     *int       `json:"read_delay_ms,omitempty"`
	RetryDelayMs    *int       `json:"retry_delay_ms,omitempty"`
	MaxRetries      *int       `json:"max_retries,omitempty"`
	RetryableErrors []int      `json:"retryable_errors,omitempty"`
}

// setCredentialConfigs will add credential type fields (Owner, Token, AppAuth) to terraform providerConfiguration
func setCredentialConfigs(creds githubConfig, cnf terraform.ProviderConfiguration) (terraform.ProviderConfiguration, error) {
	if creds.Owner != nil {
		cnf[keyOwner] = *creds.Owner
	}

	if creds.Token != nil {
		cnf[keyToken] = *creds.Token
	}

	if creds.AppAuth != nil {
		if creds.Owner == nil {
			return cnf, errors.Errorf(errTerraformProviderMissingOwner)
		}

		aaList := []map[string]any{}

		aa := map[string]any{
			keyAppAuthID:             (*creds.AppAuth)[0].ID,
			keyAppAuthInstallationID: (*creds.AppAuth)[0].InstallationID,
			keyAppAuthPemFile:        (*creds.AppAuth)[0].AuthPemFile,
		}

		aaList = append(aaList, aa)
		cnf[keyAppAuth] = aaList
	}

	return cnf, nil
}

// setParameterConfigs will add configuration type fields (WriteDelayMs, ReadDelayMs, RetryDelayMs, MaxRetries, RetryableErrors) to terraform providerConfiguration
func setParameterConfigs(creds githubConfig, cnf terraform.ProviderConfiguration) terraform.ProviderConfiguration {
	if creds.WriteDelayMs != nil {
		cnf[keyWriteDelayMs] = *creds.WriteDelayMs
	}

	if creds.ReadDelayMs != nil {
		cnf[keyReadDelayMs] = *creds.ReadDelayMs
	}

	if creds.RetryDelayMs != nil {
		cnf[keyRetryDelayMs] = *creds.RetryDelayMs
	}

	if creds.MaxRetries != nil {
		cnf[keyMaxRetries] = *creds.MaxRetries
	}

	if creds.RetryableErrors != nil {
		cnf[keyRetryableErrors] = creds.RetryableErrors
	}

	return cnf
}

func terraformProviderConfigurationBuilder(creds githubConfig) (terraform.ProviderConfiguration, error) {
	cnf := terraform.ProviderConfiguration{}

	if creds.BaseURL != nil {
		cnf[keyBaseURL] = *creds.BaseURL
	}

	cnf, err := setCredentialConfigs(creds, cnf)
	if err != nil {
		return cnf, errors.Errorf(errTerraformProviderMissingOwner)
	}

	cnf = setParameterConfigs(creds, cnf)

	return cnf, nil
}

// The terraform provider currently doesn't refresh installation tokens automatically
// Therefore, the terraform provider config needs to be refreshed at least every hour
// Once this PR is merged to terraform provider, the cache expiry can be removed
// https://github.com/integrations/terraform-provider-github/pull/2695

type CachedTerraformSetup struct {
	setup  *terraform.Setup
	expiry time.Time
}

const (
	tfSetupCacheTTL = time.Minute * 55
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which returns Terraform provider setup configuration
//
//gocyclo:ignore
func TerraformSetupBuilder(tfProvider *schema.Provider, l logging.Logger) terraform.SetupFn {
	var tfSetupLock sync.RWMutex
	tfSetups := make(map[string]CachedTerraformSetup)
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}

		l.Debug("Locking in order to update credentials")
		ok := tfSetupLock.TryLock()
		if !ok {
			return ps, errors.New(errLockError)
		}
		l.Debug("Lock succedeed")
		defer unlockMutex(&tfSetupLock, l)

		tfSetup, ok := tfSetups[configRef.Name]
		if ok && tfSetup.expiry.After(time.Now()) {
			return *tfSetup.setup, nil
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
		if data != nil {
			if err := json.Unmarshal(data, &creds); err != nil {
				return ps, errors.Wrap(err, errUnmarshalCredentials)
			}
		}

		ps.Configuration, err = terraformProviderConfigurationBuilder(creds)
		if err != nil {
			return ps, errors.Wrap(err, errProviderConfigurationBuilder)
		}

		err = configureNoForkGithubClient(ctx, &ps, *tfProvider)
		if err != nil {
			return ps, errors.Wrap(err, "failed to configure the Terraform Github provider meta")
		}

		tfSetups[configRef.Name] = CachedTerraformSetup{
			setup:  &ps,
			expiry: time.Now().Add(tfSetupCacheTTL),
		}

		return ps, nil
	}
}

func unlockMutex(lock *sync.RWMutex, l logging.Logger) {
	l.Debug("Initiating unlock")
	lock.Unlock()
	l.Debug("Unlock succeeded")
}

func configureNoForkGithubClient(ctx context.Context, ps *terraform.Setup, p schema.Provider) error {
	// Please be aware that this implementation relies on the schema.Provider
	// parameter `p` being a non-pointer. This is because normally
	// the Terraform plugin SDK normally configures the provider
	// only once and using a pointer argument here will cause
	// race conditions between resources referring to different
	// ProviderConfigs.
	diag := p.Configure(context.WithoutCancel(ctx), &tfsdk.ResourceConfig{
		Config: ps.Configuration,
	})
	if diag != nil && diag.HasError() {
		return errors.Errorf("failed to configure the provider: %v", diag)
	}
	ps.Meta = p.Meta()
	return nil
}
