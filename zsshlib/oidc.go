package zsshlib

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func OIDCFlow(initialContext context.Context, flags *SshFlags) (string, error) {
	callbackPath := "/auth/callback"
	cfg := &OIDCConfig{
		Config: oauth2.Config{
			ClientID:     flags.OIDC.ClientID,
			ClientSecret: flags.OIDC.ClientSecret,
			RedirectURL:  fmt.Sprintf("http://localhost:%v%v", flags.OIDC.CallbackPort, callbackPath),
		},
		CallbackPath: callbackPath,
		CallbackPort: flags.OIDC.CallbackPort,
		Issuer:       flags.OIDC.Issuer,
		Logf:         log.Debugf,
	}
	waitFor := 30 * time.Second
	ctx, cancel := context.WithTimeout(initialContext, waitFor)
	defer cancel() // Ensure the cancel function is called to release resources

	log.Infof("OIDC requested. If the CLI appears to be hung, check your browser for a login prompt. Waiting up to %v", waitFor)
	token, err := GetToken(ctx, cfg)
	if err != nil {
		return "", err
	}

	log.Debugf("ID token: %s", token)
	log.Infof("OIDC auth flow succeeded")

	return token, nil
}
