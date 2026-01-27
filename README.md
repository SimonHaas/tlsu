# tlsu

TODO https://certbot-dns-cloudflare.readthedocs.io/en/stable/
TODO https://hub.docker.com/r/certbot/dns-cloudflare

docker run -it --rm \
            -v "./etc/letsencrypt:/etc/letsencrypt" \
            -v "./var/lib/letsencrypt:/var/lib/letsencrypt" \
            certbot/certbot certonly --manual --agree-tos --preferred-challenges dns-01 --server https://acme-staging-v02.api.letsencrypt.org/directory -m tlsu@simonhaas.eu -d "*.umbrel.simonhaas.eu"

sudo chown -R codespace:codespace etc/ var/

docker network ls
docker network create umbrel_main_network
docker inspect   -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' tlsu-traefik-1

docker rm -f $(docker ps -aq)

git clone https://github.com/getumbrel/umbrel.git
git clone --depth 1 --branch 1.4.0 https://github.com/getumbrel/umbrel.git
sudo rm -r umbrel/.git

git clone https://github.com/dockur/umbrel.git dockerr/umbrel
sudo rm -r dockerr/umbrel/.git

cd dockerr/umbrel/
docker build --build-arg VERSION_ARG=1.4.0 . -t umbrel-self
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock -p 8889:80 umbrel-self

Mein Patch ist unter 
/workspaces/tlsu/umbrel/packages/ui/src/hooks/use-launch-app.ts
mit 
			/*** tlsu ***/
markiert.

umbrel dev instance
cd umbrel
npm run dev
Because here we are inside a devcontainer, we have to modify the volume mount in scripts/umbrel-dev.
We can not modify the environment variable PWD because it has special meaning for bash.
npm run dev logs

https://doc.traefik.io/traefik/reference/routing-configuration/http/routing/rules-and-priority/#path-pathprefix-and-pathregexp
