package azure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// MockScheme is the config URI prefix that selects a MockClient instead of
// a real azsecrets-backed Client. A vault configured with this scheme
// (e.g. "mock://local") needs no Azure credential and never leaves the
// process.
const MockScheme = "mock://"

type mockVersion struct {
	version string
	value   string
	created time.Time
	enabled bool
}

// MockClient is an in-memory fake vault for local development and testing,
// seeded with a handful of realistic-looking secrets. It never touches the
// network or the filesystem.
type MockClient struct {
	mu      sync.Mutex
	secrets map[string][]mockVersion // each slice ordered oldest -> newest
}

var _ SecretsClient = (*MockClient)(nil)

// NewMockClient returns a MockClient pre-seeded with sample secrets,
// including one with multiple versions so the version-history panel has
// something to show.
func NewMockClient() *MockClient {
	c := &MockClient{secrets: make(map[string][]mockVersion)}
	now := time.Now()

	// These are placeholder strings, not real credentials — deliberately
	// avoid formats real secret scanners key off of (no sk_/whsec_/AKIA-style
	// prefixes, no realistic base64 blobs), so they don't get mistaken for a
	// leak.
	c.seed("database-connection-string", "MOCK-DB-CONN not-a-real-secret rev1", now.Add(-72*time.Hour))
	c.seed("database-connection-string", "MOCK-DB-CONN not-a-real-secret rev2", now.Add(-2*time.Hour))

	c.seed("api-signing-key", "MOCK-API-SIGNING-KEY not-a-real-secret", now.Add(-200*time.Hour))

	c.seed("storage-account-key", "MOCK-STORAGE-ACCOUNT-KEY not-a-real-secret", now.Add(-500*time.Hour))

	c.seed("smtp-password", "MOCK-SMTP-PASSWORD not-a-real-secret rev1", now.Add(-24*time.Hour))
	c.seed("smtp-password", "MOCK-SMTP-PASSWORD not-a-real-secret rev2", now.Add(-1*time.Hour))

	c.seed("third-party-webhook-secret", "MOCK-WEBHOOK-SECRET not-a-real-secret", now.Add(-10*time.Hour))

	return c
}

func (c *MockClient) seed(name, value string, created time.Time) {
	c.secrets[name] = append(c.secrets[name], mockVersion{
		version: randomVersionID(),
		value:   value,
		created: created,
		enabled: true,
	})
}

func randomVersionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%032d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func (c *MockClient) ListSecretNames(ctx context.Context) ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	names := make([]string, 0, len(c.secrets))
	for name := range c.secrets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func (c *MockClient) GetSecret(ctx context.Context, name, version string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	versions, ok := c.secrets[name]
	if !ok || len(versions) == 0 {
		return "", fmt.Errorf("secret %q not found", name)
	}
	if version == "" {
		return versions[len(versions)-1].value, nil
	}
	for _, v := range versions {
		if v.version == version {
			return v.value, nil
		}
	}
	return "", fmt.Errorf("version %q of secret %q not found", version, name)
}

func (c *MockClient) SetSecret(ctx context.Context, name, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.secrets[name]; !ok {
		return fmt.Errorf("secret %q not found", name)
	}
	c.secrets[name] = append(c.secrets[name], mockVersion{
		version: randomVersionID(),
		value:   value,
		created: time.Now(),
		enabled: true,
	})
	return nil
}

func (c *MockClient) ListSecretVersions(ctx context.Context, name string) ([]Version, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	versions, ok := c.secrets[name]
	if !ok {
		return nil, fmt.Errorf("secret %q not found", name)
	}
	out := make([]Version, len(versions))
	for i, v := range versions {
		out[i] = Version{
			Version: v.version,
			Enabled: v.enabled,
			Created: v.created.Format("2006-01-02 15:04:05"),
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Created > out[j].Created
	})
	return out, nil
}
