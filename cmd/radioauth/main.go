package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/holoplot/sw__radioauth/account"
	"golang.org/x/oauth2"
)

type configFile struct {
	OAuthClientID     string   `json:"oauth_client_id,omitempty"`
	OAuthClientSecret string   `json:"oauth_client_secret,omitempty"`
	OAuthIssuer       string   `json:"oauth_issuer,omitempty"`
	OAuthCallbackURL  string   `json:"oauth_callback_url,omitempty"`
	OAuthAccountURL   string   `json:"oauth_account_url,omitempty"`
	AccountStorePath  string   `json:"account_store_path,omitempty"`
	RedisAddresses    []string `json:"redis_addresses,omitempty"`
	HTTPPort          uint16   `json:"http_port,omitempty"`
	RadiusSecret      string   `json:"radius_secret,omitempty"`
}

var (
	provider            *oidc.Provider
	oauthConfig         oauth2.Config
	accountStore        account.Store
	config              configFile
	relativeCallbackURL string
)

func authenticateToken(account *account.Account) bool {
	storedOauth2Token := oauth2.Token{
		AccessToken:  account.AccessToken,
		RefreshToken: account.RefreshToken,
		Expiry:       account.TokenExpiry,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tokenSource := oauthConfig.TokenSource(ctx, &storedOauth2Token)

	userinfo, err := provider.UserInfo(ctx, tokenSource)
	if err != nil {
		log.Printf("[authenticate] OAuth provider rejected %s: %v", account.Username, err)
		return false
	}

	currentOauth2Token, err := tokenSource.Token()
	if err != nil {
		log.Printf("[authenitaction] Unable to obtain OAuth Token for user %s: %v", account.Username, err)
		return false
	}

	if !currentOauth2Token.Valid() {
		log.Printf("[authenticate] OAuth token for user %s is invalid!", account.Username)
		return false
	}

	if userinfo.Email != account.Username {
		log.Printf("[authenticate] OAuth token belongs to %s, not %s, rejecting", userinfo.Email, account.Username)
		return false
	}

	if !userinfo.EmailVerified {
		log.Printf("[authenticate] Email address of %s not verified, rejecting", userinfo.Email)
		return false
	}

	// Sync back the (possibly refreshed) access token
	account.AccessToken = currentOauth2Token.AccessToken
	account.RefreshToken = currentOauth2Token.RefreshToken
	account.TokenExpiry = currentOauth2Token.Expiry

	err = accountStore.Write(account)
	if err != nil {
		log.Printf("[authenticate] Cannot write back account info for %s: %v", userinfo.Email, err)
		return false
	}

	return true
}

func main() {
	configPathFlag := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	b, err := ioutil.ReadFile(*configPathFlag)
	if err != nil {
		log.Fatalf("Cannot read config file: %v\n", err)
		return
	}

	err = json.Unmarshal(b, &config)

	if err != nil {
		log.Fatalf("Cannot parse config file: %v\n", err)
		return
	}

	if len(config.AccountStorePath) > 0 {
		accountStore, err = account.NewFileStore(config.AccountStorePath)
		if err != nil {
			log.Fatalf("Cannot create file-backed account store: %v\n", err)
		}
		log.Printf("Using file-backed account store at path %s\n", config.AccountStorePath)
	} else if len(config.RedisAddresses) > 0 {
		accountStore, err = account.NewRedisStore(config.RedisAddresses)
		if err != nil {
			log.Fatalf("Cannot create redis-backed account store: %v\n", err)
		}
		log.Printf("Using redis-backed account store with addresses %v\n", config.RedisAddresses)
	} else {
		log.Fatalf("No account store available")
	}

	ctx := context.Background()

	provider, err = oidc.NewProvider(ctx, config.OAuthIssuer)
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse(config.OAuthCallbackURL)
	if err != nil {
		log.Fatal(err)
	}

	relativeCallbackURL = u.Path

	oauthConfig = oauth2.Config{
		ClientID:     config.OAuthClientID,
		ClientSecret: config.OAuthClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  config.OAuthCallbackURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	go runHTTPServer()
	go runRadiusServer()

	for {
		select {}
	}
}
