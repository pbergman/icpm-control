docker-build-mipsel:
	sudo docker build --tag mipsel-build-env  --file Dockerfile.mipsel .

docker-start-mipsel: docker-build-mipsel
	sudo docker run \
		--detach \
		--rm \
		--tty \
		--interactive \
		--env GOOS=linux \
		--env GOARCH=mipsle \
		--env CGO_ENABLED=1 \
		--env CC=mipsel-linux-gnu-gcc \
		--env GOROOT=/usr/local/go \
		--env PATH=$$GOROOT/bin:$$PATH \
		--mount type=bind,source="$$(pwd)",target=/app \
		--workdir /app \
		--entrypoint bash \
		--name mipsel-build-env \
	  	mipsel-build-env

docker-stop-mipsel:
	sudo docker stop mipsel-build-env

build-server-mipsel:
	if [ ! -d "./build/server-mipsel" ]; then mkdir -p "./build/server-mipsel"; fi
	sudo docker exec -it mipsel-build-env /usr/local/go/bin/go build -o ./build/server-mipsel/icpm-control ./server
	sudo mips-linux-gnu-strip ./build/server-mipsel/icpm-control

build-client:
	if [ ! -d "./build/client" ]; then mkdir -p "./build/client"; fi
	sudo go build -o ./build/client/icpm-control ./client
