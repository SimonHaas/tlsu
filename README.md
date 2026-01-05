# tlsu

TODO https://certbot-dns-cloudflare.readthedocs.io/en/stable/
TODO https://hub.docker.com/r/certbot/dns-cloudflare

docker run -it --rm \
            -v "./etc/letsencrypt:/etc/letsencrypt" \
            -v "./var/lib/letsencrypt:/var/lib/letsencrypt" \
            certbot/certbot certonly --manual --agree-tos --preferred-challenges dns-01 --server https://acme-staging-v02.api.letsencrypt.org/directory -m tlsu@simonhaas.eu -d "*.umbrel.simonhaas.eu"

sudo chown -R codespace:codespace etc/ var/

