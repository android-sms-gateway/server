# Building the binary of the App
FROM golang:1.22-alpine AS build

ARG APP
ARG APP_VERSION=1.0.0
ARG APP_RELEASE_ID=1
WORKDIR /go/src

# Install swag
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go.mod and go.sum
COPY go.* ./

# Downloads all the dependencies in advance (could be left out, but it's more clear this way)
RUN go mod download

# Copy all the Code and stuff to compile everything
COPY . .

# Replace {APP_VERSION} with the actual version
RUN sed -i "s/{APP_VERSION}/${APP_VERSION}/g" ./cmd/${APP}/main.go

# Generate swagger docs
RUN go generate ./...

# Builds the application as a staticly linked one, to allow it to run on alpine
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s -X github.com/android-sms-gateway/server/internal/version.AppVersion=${APP_VERSION} -X github.com/android-sms-gateway/server/internal/version.AppRelease=${APP_RELEASE_ID}" -o app ./cmd/${APP}/main.go

# Moving the binary to the 'final Image' to make it smaller
FROM alpine:3 AS prod

WORKDIR /app

RUN apk add --no-cache tzdata

COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh

COPY --from=build /go/src/app /app

# Exposes port 3000 because our program listens on that port
EXPOSE 3000

USER guest
HEALTHCHECK --interval=10s --timeout=3s --retries=3 --start-period=5s CMD /app/app health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
