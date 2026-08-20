package azure

import (
	"context"
	"sort"
	"testing"
)

func TestMockClient_ListSecretNames(t *testing.T) {
	c := NewMockClient()

	names, err := c.ListSecretNames(context.Background())
	if err != nil {
		t.Fatalf("ListSecretNames() unexpected error: %v", err)
	}

	want := []string{
		"api-signing-key",
		"database-connection-string",
		"smtp-password",
		"storage-account-key",
		"third-party-webhook-secret",
	}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("ListSecretNames() = %v, want sorted", names)
	}
	if len(names) != len(want) {
		t.Fatalf("ListSecretNames() returned %d names, want %d: %v", len(names), len(want), names)
	}
	for i, n := range want {
		if names[i] != n {
			t.Fatalf("ListSecretNames()[%d] = %q, want %q", i, names[i], n)
		}
	}
}

func TestMockClient_GetSecret(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	t.Run("current version with empty version arg", func(t *testing.T) {
		value, err := c.GetSecret(ctx, "smtp-password", "")
		if err != nil {
			t.Fatalf("GetSecret() unexpected error: %v", err)
		}
		if value != "MOCK-SMTP-PASSWORD not-a-real-secret rev2" {
			t.Fatalf("GetSecret() = %q, want the newest seeded revision", value)
		}
	})

	t.Run("specific historical version", func(t *testing.T) {
		versions, err := c.ListSecretVersions(ctx, "smtp-password")
		if err != nil {
			t.Fatalf("ListSecretVersions() unexpected error: %v", err)
		}
		if len(versions) != 2 {
			t.Fatalf("ListSecretVersions() returned %d versions, want 2", len(versions))
		}
		// versions are newest-first; the oldest one is rev1.
		oldest := versions[len(versions)-1]
		value, err := c.GetSecret(ctx, "smtp-password", oldest.Version)
		if err != nil {
			t.Fatalf("GetSecret() unexpected error: %v", err)
		}
		if value != "MOCK-SMTP-PASSWORD not-a-real-secret rev1" {
			t.Fatalf("GetSecret() = %q, want rev1", value)
		}
	})

	t.Run("unknown secret", func(t *testing.T) {
		_, err := c.GetSecret(ctx, "does-not-exist", "")
		if err == nil {
			t.Fatal("GetSecret() expected error for unknown secret, got nil")
		}
	})

	t.Run("unknown version", func(t *testing.T) {
		_, err := c.GetSecret(ctx, "smtp-password", "not-a-real-version-id")
		if err == nil {
			t.Fatal("GetSecret() expected error for unknown version, got nil")
		}
	})
}

func TestMockClient_SetSecret(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	t.Run("appends a new current version", func(t *testing.T) {
		before, err := c.ListSecretVersions(ctx, "api-signing-key")
		if err != nil {
			t.Fatalf("ListSecretVersions() unexpected error: %v", err)
		}

		if err := c.SetSecret(ctx, "api-signing-key", "MOCK-API-SIGNING-KEY updated"); err != nil {
			t.Fatalf("SetSecret() unexpected error: %v", err)
		}

		after, err := c.ListSecretVersions(ctx, "api-signing-key")
		if err != nil {
			t.Fatalf("ListSecretVersions() unexpected error: %v", err)
		}
		if len(after) != len(before)+1 {
			t.Fatalf("ListSecretVersions() returned %d versions after SetSecret, want %d", len(after), len(before)+1)
		}

		value, err := c.GetSecret(ctx, "api-signing-key", "")
		if err != nil {
			t.Fatalf("GetSecret() unexpected error: %v", err)
		}
		if value != "MOCK-API-SIGNING-KEY updated" {
			t.Fatalf("GetSecret() = %q, want the newly set value", value)
		}
	})

	t.Run("unknown secret", func(t *testing.T) {
		if err := c.SetSecret(ctx, "does-not-exist", "value"); err == nil {
			t.Fatal("SetSecret() expected error for unknown secret, got nil")
		}
	})
}

func TestMockClient_ListSecretVersions_NewestFirst(t *testing.T) {
	c := NewMockClient()
	ctx := context.Background()

	versions, err := c.ListSecretVersions(ctx, "database-connection-string")
	if err != nil {
		t.Fatalf("ListSecretVersions() unexpected error: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("ListSecretVersions() returned %d versions, want 2", len(versions))
	}
	if versions[0].Created < versions[1].Created {
		t.Fatalf("ListSecretVersions() not newest-first: %v", versions)
	}
}

func TestMockClient_ListSecretVersions_UnknownSecret(t *testing.T) {
	c := NewMockClient()
	if _, err := c.ListSecretVersions(context.Background(), "does-not-exist"); err == nil {
		t.Fatal("ListSecretVersions() expected error for unknown secret, got nil")
	}
}

func TestMockClient_ImplementsSecretsClient(t *testing.T) {
	var _ SecretsClient = NewMockClient()
}
