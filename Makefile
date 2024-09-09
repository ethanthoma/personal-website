IMAGE_NAME = webserver
TAG = 0.1

run: build
	sudo docker load < result
	sudo docker run --rm --env-file ./.env -p 127.0.0.1:8080:8080 -t $(IMAGE_NAME):$(TAG)

build:
	nix build .#container