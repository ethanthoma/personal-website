IMAGE_NAME = webserver
TAG = 0.1
PORT = ${WEBSERVER_PORT}
SYSTEM = x86_64-linux

run:
	nix run github:Mic92/nix-fast-build -- --flake '.#packages.$(SYSTEM).default'
	nix run

docker/build:
	nix build .#container

docker: docker/build
	docker load < result
	docker run \
		--rm \
		--env-file ./.env \
		-p 127.0.0.1:$(PORT):$(PORT) \
		-t $(IMAGE_NAME):$(TAG)

live/templ:
	templ generate \
		--watch \
		--proxy="http://localhost:3001" \
		--open-browser=false \
		-v \
		--path=./services/webserver/

live/server:
	air \
		--build.cmd "go build -o tmp/bin/main personal-website/services/webserver" \
		--build.bin "tmp/bin/main" \
		--build.delay "100" \
		--build.include_dir "personal-website/services/webserver" \
		--build.include_ext "go" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true \
		--proxy.enabled true \
		--proxy.proxy_port 3001  \
		--proxy.app_port $(PORT)

live/tailwind:
	tailwindcss \
		-c ./services/webserver/tailwind.config.js \
		-i ./services/webserver/public/main.css \
		-o ./public/main.css \
		-m \
		-w

live/sync_assets:
	rsync -a ./services/webserver/public/* ./public --exclude='*.css'
	sleep 0.1 
	air \
		--build.cmd "templ generate --notify-proxy" \
		--build.bin "true" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_dir "public" \
		--build.include_ext "js,css"

live: 
	make live/templ & \
	make live/server & \
	make live/tailwind & \
	make live/sync_assets
