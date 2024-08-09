## Overview

This directory implements authentication and authorization for Veraison API.
Authentication can be performed using the Basic HTTP scheme (with the `basic`
backend), or using a Bearer token (with the `keycloak` backend). Once an API
user is authenticated, authorization is
[role-based](https://en.wikipedia.org/wiki/Role-based_access_control). See
documentation for specific services for which role(s) are needed to access
their API.


## Configuration

- `backend`: specifies which auth backend will be used by the service. The
  valid options are:

  - `passthrough`: a backend that does not perform any authentication, allowing
    all requests.
  - `none`: alias for `passthrough`.
  - `basic`: Uses the Basic HTTP authentication scheme. See
    [RFC7617](https://datatracker.ietf.org/doc/html/rfc7617) for details. This
    is not intended for production.
  - `keycloak`: Uses OpenID Connect protocol as implemented by the Keycloak
    authentication server.

  See below for details of how to configure individual backends.

### Passthrough

No additional configuration is required. `passthrough` will allow all requests.
This is the default if `auth` configuration is not provided.

### Basic

- `users`: this is a mapping of user names onto their bcrypt password hashes
  and roles. The key of the mapping is the user name, the value is a further
  mapping for the details with the following fields:

    - `password`: the bcrypt hash of the user's password.
    - `roles`: either a single role or a list of roles associated with the
      user. API authrization will be performed based on the user's roles.

On Linux, bcrypt hashes can be generated on the command line using `mkpasswd`
utility, e.g.:

```bash
mkpasswd -m bcrypt --stdin <<< Passw0rd!
```

For example:

```yaml
auth:
  backend: basic
  users:
    user1:
      password: "$2b$05$XgVBveh6QPrRHXI.8S/J9uobBR7Wv9z4CL8yACHEmKIQmYSSyKAqC" # Passw0rd!
      roles: provisioner
    user2:
      password: "$2b$05$x5fvAV5WPkX0KXzqf5FMKODz0uyi2ioew1lOrF2Czp2aNH1LQmhki" # @s3cr3t
      roles: [manager, provisioner]
```

### Keycloak

- `host` (optional): host name of the Keycloak service. Defaults to
  `localhost`.
- `port` (optional): the port on which the Keycloak service is listening.
  Defaults to `8080`.
- `realm` (optional): the Keycloak realm used by Veraison. A realm contains the
  configuration for clients, users, roles, etc. It is roughly analogous to a
  "tenant id". Defaults to `veraison`.
- `ca-cert`: the path to a PEM-encoded x509 cert that will be added to CA certs
  when establishing connection to the Keycloak server. This should be specified
  if the server has HTTPS enabled and the root CA for its cert is not installed
  in the system.

For example:

```yaml
auth:
  backend: keycloak
  host: keycloak.example.com
  port: 11111
  realm: veraison
```

## Usage

```go
"github.com/gin-gonic/gin"
"github.com/veraison/services/auth"
"github.com/veraison/services/config"
"github.com/veraison/services/log"

func main() {
    // Load authroizer config.
	v, err := config.ReadRawConfig(*config.File, false)
	if err != nil {
		log.Fatalf("Could not read config: %v", err)
	}
	subs, err := config.GetSubs(v, "*auth")
	if err != nil {
		log.Fatalf("Could not parse config: %v", err)
	}

    // Create new authorizer based on the loaded config.
	authorizer, err := auth.NewAuthorizer(subs["auth"], log.Named("auth"))
	if err != nil {
		log.Fatalf("could not init authorizer: %v", err)
	}

    // Ensure the authorizer is terminated properly on exit
	defer func() {
		err := authorizer.Close()
		if err != nil {
			log.Errorf("Could not close authorizer: %v", err)
		}
	}()

    // Use the authorizer to set a middleware handler in the appropriate gin
    // router, with an appropriate role.
    router := gin.Default()
	router.Use(authorizer.GetGinHandler(auth.ManagerRole))

    // Set up route handling here
    // ...
    // ...

    // Run the service.
    if err := router.Run("0.0.0.0:80"); err != nil {
		log.Errorf("Gin engine failed: %v", err)
    }
}

```

