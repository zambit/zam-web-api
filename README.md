# ZAM Wallet-Web-Api

This project exposes web-server which is the part of Wallet-Api

## Installation

### Requirements

* Configured env with Go >= 1.10
* Installed [dep](https://github.com/golang/dep) utility
* Installed [migrate](https://github.com/golang-migrate/migrate) utility
* Postgresql database
* Installer [ginkgo](https://github.com/onsi/ginkgo) utility (for testes only)

Assumed that all commands are invoked in the root on this project.

### Dependencies

Before build it's required to populate all dependencies, just execute

```bash
dep ensure
```

and wait until complete.

### Testing

Execute in bash

```bash
ginkgo -r .
```

Currently tests assumes that database for tests are accessible via uri `postgresql://test:test@localhost/test?sslmode=disable`. Also tests requires clean DB and runs migrations internally.

### Building

Execute in bash

```bash
go build -o {executable_name} cmd/main/main.go
```

It will produces statically linked executable which depends only on `libc`.

### Migrations

Migrations are implemented via [migrate](https://github.com/golang-migrate/migrate) utility.

Basically last revision can be applied by executing

```bash
migrate -path=db/migrations -database=${YOUR_PORSTRES_URI} up
```

### Configuration

Example configuration file which describes scheme and comments each value

```yaml
# Env describes current environment
env: test
# DB connection description
db:
  # URI contains all necessary connection parts in URI form
  # Described here https://www.postgresql.org/docs/current/static/libpq-connect.html#id-1.7.3.8.3.2.
  uri: postgresql://postgres:postgres@localhost:5432/postgres

# Server holds different web-server related configuration values
server:
  # Host to listen on such address, accept both ip4 and ip6 addresses
  host: localhost
  # Port to listen on, negative values will cause UB
  port: 9999
  # Web-authorization related parameters
  auth:
    # Specifies token prefix in Authorization header
    tokenname: Bearer
    # Authorization token live duration before become expire (example: 24h45m15s)
    tokenexpire: 24h0m0s
```

## Running

### Server

Represents command which runs **Wallet-Web-Api** server, it accept arguments which overrides values specified by configuration.

#### Usage

```bash
./wallet_api_binary [root_flags] server [flags]
```

#### Root flags

```
-c, --config string   specifies configuration file to load from
-e, --env string      specifies current environment (prod/dev/test) (default "test")
```

#### Flags

```
    --db.uri string        postgres connection uri (default "postgresql://postgres:postgres@localhost:5432/postgres")
-l, --server.host string   host to serve on (default "localhost")
-p, --server.port int      port to serve on (default 9999)
```

## Exported endpoints

Currently **Wallet-Web-Api** exports such endpoints

* `POST   /api/v1/auth/signup/start`
* `POST   /api/v1/auth/signup/verify`
* `PUT    /api/v1/auth/signup/finish`
* `POST   /api/v1/auth/recovery/finish`
* `POST   /api/v1/auth/recovery/finish`
* `PUT    /api/v1/auth/recovery/finish`
* `POST   /api/v1/auth/signin`
* `DELETE /api/v1/auth/signout`
* `GET    /api/v1/auth/check`
* `POST   /api/v1/auth/refresh_token`

Also some endpoints requires `Authorization` header, so it have not be filtered.
