package app

import "github.com/NithishKumar2772/TBox-Drive/internal/security/keyring"

// NewOSSecretStore uses the platform credential vault for OAuth secrets.
// On Linux, a Secret Service provider must be available in the desktop session.
func NewOSSecretStore()(SecretStore,error){return keyring.New("TBox Drive")}
