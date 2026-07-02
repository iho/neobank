package runtime

import "fmt"

const defaultJWTSecret = "dev-secret-change-me"

// RequireProductionSecrets rejects default or missing JWT when running in production.
func RequireProductionSecrets(appEnv, jwtSecret string) error {
	switch appEnv {
	case "production", "prod":
	default:
		return nil
	}
	if jwtSecret == "" || jwtSecret == defaultJWTSecret {
		return fmt.Errorf("JWT_SECRET must be set to a non-default value when APP_ENV=%s", appEnv)
	}
	return nil
}