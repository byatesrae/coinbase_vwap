ARG BUILD_IMAGE=golang
FROM $BUILD_IMAGE:1.19.2-alpine3.16

WORKDIR /opt/app

CMD ["go", "run", "/opt/app/cmd/coinbasevwap/"]
