# The build/ci environment for this project.
ARG BUILD_IMAGE=golang
FROM $BUILD_IMAGE:1.19.2-alpine3.16

RUN apk add --no-cache --upgrade \
    bash \
    curl \
    make \
    g++ \
    git

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1

RUN go install github.com/matryer/moq@v0.3.0