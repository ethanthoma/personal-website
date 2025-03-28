IMAGE_NAME = webserver
TAG = 0.1
PORT = ${WEBSERVER_PORT}
SYSTEM = x86_64-linux

run:
	@nix run github:Mic92/nix-fast-build -- --flake '.#packages.$(SYSTEM).default'
	@nix run

docker/build:
	@nix build .#container

docker: build/docker
	@docker load < result
	@docker run \
		--rm \
		--env-file ./.env \
		-p 127.0.0.1:$(PORT):$(PORT) \
		-t $(IMAGE_NAME):$(TAG)

live/templ:
	@templ generate \
		--watch \
		--proxy="http://localhost:$(PORT)" \
		--open-browser=false \
		-v \
		--path=./services/webserver/

live/server:
	@air \
		--build.cmd "go build -o tmp/bin/main personal-website/services/webserver" \
		--build.bin "tmp/bin/main" \
		--build.delay "100" \
		--build.include_ext "go" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true \
		--proxy.enabled true

live/tailwind:
	echo "running tailwind"
	@tailwindcss \
		-i ./services/webserver/public/main.css \
		-o ./public/main.css \
		--watch=always \
		-cwd=./services/webserver
	echo "finished... tailwind"

live/sync_assets:
	@sleep 1000
	@air \
		--build.cmd "templ generate --notify-proxy" \
		--build.bin "rsync -a ./services/webserver/public/* ./public --exclude='*.css'" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_dir "./public" \
		--build.include_ext "js,css"

live: 
	make -j4 \
		live/server \
		live/templ \
		live/tailwind \
		live/sync_assets
