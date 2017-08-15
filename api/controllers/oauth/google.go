package oauth

import (
	"github.com/owen-d/beacon-api/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func NewOAuthConf(vars *config.OAuth, env string) *oauth2.Config {
	var redirectUri string
	if env == "production" {
		redirectUri = vars.RedirectUris.Prod
	} else {
		redirectUri = vars.RedirectUris.Dev
	}
	return &oauth2.Config{
		ClientID:     vars.ClientID,
		ClientSecret: vars.ClientSecret,
		RedirectURL:  redirectUri,
		Scopes:       vars.Scopes,
		Endpoint:     google.Endpoint,
	}
}

type GoogleAuth interface{}

type GoogleAuthMethods struct {
	OAuth *oauth2.Config
}
