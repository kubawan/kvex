// Package azure is a thin wrapper around azsecrets.Client exposing only the
// operations kvex needs: listing secret names, fetching a value (optionally
// at a specific version), listing version history, and setting a new value.
package azure

import (
	"context"
	"fmt"
	"sort"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// SecretsClient is the set of vault operations the UI depends on. The real
// Client (backed by azsecrets) and MockClient (an in-memory fake for local
// testing without Azure access) both implement it.
type SecretsClient interface {
	ListSecretNames(ctx context.Context) ([]string, error)
	GetSecret(ctx context.Context, name, version string) (string, error)
	SetSecret(ctx context.Context, name, value string) error
	ListSecretVersions(ctx context.Context, name string) ([]Version, error)
}

// Client wraps a single vault's azsecrets.Client.
type Client struct {
	secrets *azsecrets.Client
}

var _ SecretsClient = (*Client)(nil)

// NewClient creates a Client for the given vault URI using the supplied
// credential (a single azidentity.DefaultAzureCredential is shared across
// all vaults).
func NewClient(vaultURI string, credential azcore.TokenCredential) (*Client, error) {
	c, err := azsecrets.NewClient(vaultURI, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("creating client for %s: %w", vaultURI, err)
	}
	return &Client{secrets: c}, nil
}

// ListSecretNames returns the names of every enabled-or-not secret in the
// vault, sorted alphabetically. It never fetches secret values.
func (c *Client) ListSecretNames(ctx context.Context) ([]string, error) {
	var names []string
	pager := c.secrets.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing secrets: %w", err)
		}
		for _, item := range page.Value {
			if item.ID == nil {
				continue
			}
			names = append(names, item.ID.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// GetSecret fetches a secret's value. Pass an empty version to get the
// current version.
func (c *Client) GetSecret(ctx context.Context, name, version string) (string, error) {
	resp, err := c.secrets.GetSecret(ctx, name, version, nil)
	if err != nil {
		return "", err
	}
	if resp.Value == nil {
		return "", nil
	}
	return *resp.Value, nil
}

// SetSecret creates a new version of an existing secret with the given value.
func (c *Client) SetSecret(ctx context.Context, name, value string) error {
	_, err := c.secrets.SetSecret(ctx, name, azsecrets.SetSecretParameters{
		Value: &value,
	}, nil)
	if err != nil {
		return fmt.Errorf("setting secret %s: %w", name, err)
	}
	return nil
}

// Version describes one historical version of a secret (metadata only).
type Version struct {
	Version string
	Enabled bool
	Created string // RFC3339, empty if unknown
}

// ListSecretVersions returns a secret's version history, newest first.
func (c *Client) ListSecretVersions(ctx context.Context, name string) ([]Version, error) {
	var versions []Version
	pager := c.secrets.NewListSecretPropertiesVersionsPager(name, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing versions of %s: %w", name, err)
		}
		for _, item := range page.Value {
			v := Version{}
			if item.ID != nil {
				v.Version = item.ID.Version()
			}
			if item.Attributes != nil {
				if item.Attributes.Enabled != nil {
					v.Enabled = *item.Attributes.Enabled
				}
				if item.Attributes.Created != nil {
					v.Created = item.Attributes.Created.Format("2006-01-02 15:04:05")
				}
			}
			versions = append(versions, v)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Created > versions[j].Created
	})
	return versions, nil
}
