# RadiOauth

RadiOauth is a Radius server that uses an OAuth/OpenID access provider to authenticate Radius clients.
It can be used in organizations that have an OAuth/OpenID infrastructure such as Google SSO in order
to allow their users access to VPN or WLAN services which facilitate Radius authentication.

To get access, users are supposed to visit the https-wrapped URL the http service in this project provides.
The code in this project does not handle any TLS endpoints, so administrators are required to reverse-proxy
the service in order to have an https endpoint.
In case a user does not have a local account yet, they will be redirected to the OAuth provider to
acknowledge the account link. Once that has succeeded, RadiOauth will create an internal account, assign a
random password to it and prompt it to the user.

The Radius server will then authenticate users with both their password as well as by asking the OAuth
provider whether the token is still valid.

## Config file

The config file is provided in JSON format and its location is passed to the server using the `-config` flag.

| Key                         | Description                                                              |
|-----------------------------|--------------------------------------------------------------------------|
| `oauth_client_id`           | The Client ID as provided by the OAuth provider                          |
| `oauth_client_secret`       | The Client Secret as provided by the OAuth provider                      |
| `oauth_issuer`              | A URL to the OAuth issuer                                                |
| `oauth_callback_url`        | The URL used by the OAuth provider for authentication callbacks. Note that this URL must be white-listed in the settings of the OAuth provider |
| `oauth_account_url`         | A URL a user can click in order to remove the App from their account     |
| `account_store_path`        | Used for file-backed account store, see below                            |
| `redis_addresses`           | Used for Redis-backed account store, see below                           |
| `radius_secret`             | A secret for the Radius server, shared with other services using it      |
| `http_port`                 | The HTTP port to listen on                                               |

## Account store

Radioauth needs an account store for user authentication. There are currently two available implementations.

### File-backed

Specify a path that is writeable by the user running the process as `account_store_path` in the config file.
This will create files with JSON content when users are created, and read from these files when users are
authentivated. It has no extra runtime dependencies.

### Redis-backed

The Redis store can be enabled by specifing one or many Redis addresses as `redis_addresses` in the config
file. When many are selected, a Redis cluster will be used. Account data is saved as JSON serialized
strings, stored with the username as key.

## Setup (Google)

To get started with Google as OAuth provider, first follow [the instruction provided on Google's Identity Platform](https://developers.google.com/identity/protocols/OpenIDConnect).

## License

MIT