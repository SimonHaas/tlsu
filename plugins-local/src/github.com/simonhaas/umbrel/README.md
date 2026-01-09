# umbrel

This should become a traefik plugin for umbrel.

https://doc.traefik.io/traefik-hub/api-gateway/guides/plugin-development-guide


# at the module root (where go.mod lives)
go mod tidy
go mod vendor
# build using vendor
go build -mod=vendor ./...

