//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vault

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/iho/neobank/pkg/piicrypto"
)

// Config for HashiCorp Vault Transit field encryption.
type Config struct {
	Addr       string
	Token      string
	TransitKey string
	HMACKey    string
}

// LoadConfigFromEnv reads Vault settings from the environment.
// Returns ok=false when VAULT_ADDR is unset (caller should use piicrypto.Noop).
func LoadConfigFromEnv() (Config, bool) {
	addr := os.Getenv("VAULT_ADDR")
	if addr == "" {
		return Config{}, false
	}
	transitKey := envOr("VAULT_TRANSIT_KEY", "pii")
	hmacKey := envOr("VAULT_HMAC_KEY", "pii-phone")
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		token = os.Getenv("VAULT_DEV_ROOT_TOKEN_ID")
	}
	return Config{
		Addr:       addr,
		Token:      token,
		TransitKey: transitKey,
		HMACKey:    hmacKey,
	}, true
}

// NewTransitProtector returns a piicrypto.Protector backed by Vault Transit.
func NewTransitProtector(cfg Config) (*TransitProtector, error) {
	client, err := api.NewClient(&api.Config{Address: cfg.Addr})
	if err != nil {
		return nil, fmt.Errorf("vault client: %w", err)
	}
	if cfg.Token != "" {
		client.SetToken(cfg.Token)
	}
	if cfg.TransitKey == "" || cfg.HMACKey == "" {
		return nil, fmt.Errorf("vault transit and hmac key names are required")
	}
	return &TransitProtector{
		client:     client,
		transitKey: cfg.TransitKey,
		hmacKey:    cfg.HMACKey,
	}, nil
}

// TransitProtector encrypts PII with Vault Transit and derives phone blind indexes via Transit HMAC.
type TransitProtector struct {
	client     *api.Client
	transitKey string
	hmacKey    string
}

func (p *TransitProtector) Enabled() bool { return true }

func (p *TransitProtector) Encrypt(ctx context.Context, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	path := fmt.Sprintf("transit/encrypt/%s", p.transitKey)
	secret, err := p.client.Logical().WriteWithContext(ctx, path, map[string]any{
		"plaintext": base64.StdEncoding.EncodeToString([]byte(plaintext)),
	})
	if err != nil {
		return "", fmt.Errorf("vault encrypt: %w", err)
	}
	ciphertext, err := secretCipher(secret)
	if err != nil {
		return "", err
	}
	return piicrypto.CiphertextPrefix + ciphertext, nil
}

func (p *TransitProtector) Decrypt(ctx context.Context, stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	if !piicrypto.IsEncrypted(stored) {
		return stored, nil
	}
	path := fmt.Sprintf("transit/decrypt/%s", p.transitKey)
	secret, err := p.client.Logical().WriteWithContext(ctx, path, map[string]any{
		"ciphertext": stored[len(piicrypto.CiphertextPrefix):],
	})
	if err != nil {
		return "", fmt.Errorf("vault decrypt: %w", err)
	}
	plainB64, err := secretDataString(secret, "plaintext")
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(plainB64)
	if err != nil {
		return "", fmt.Errorf("decode plaintext: %w", err)
	}
	return string(raw), nil
}

func (p *TransitProtector) PhoneLookup(ctx context.Context, phone string) (string, error) {
	normalized := piicrypto.NormalizePhone(phone)
	if normalized == "" {
		return "", nil
	}
	path := fmt.Sprintf("transit/hmac/%s", p.hmacKey)
	secret, err := p.client.Logical().WriteWithContext(ctx, path, map[string]any{
		"input": base64.StdEncoding.EncodeToString([]byte(normalized)),
	})
	if err != nil {
		return "", fmt.Errorf("vault hmac: %w", err)
	}
	return secretDataString(secret, "hmac")
}

func secretCipher(secret *api.Secret) (string, error) {
	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("vault: empty encrypt response")
	}
	v, ok := secret.Data["ciphertext"].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("vault: missing ciphertext")
	}
	return v, nil
}

func secretDataString(secret *api.Secret, key string) (string, error) {
	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("vault: empty response")
	}
	v, ok := secret.Data[key].(string)
	if !ok || v == "" {
		return "", fmt.Errorf("vault: missing %s", key)
	}
	return v, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}