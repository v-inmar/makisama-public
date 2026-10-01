# ----------------------------BUILD STAGE ----------------------- #
# use golang offical image but with 1.26 version
# mark it as builder
FROM golang:1.26 AS builder

# the application root directory
WORKDIR /app

# copy the go modules into current working directory ./ means current working directory
# WORKDIR must be set first, just like what we have above
COPY go.mod go.sum ./

# download the go modules that we are using
RUN go mod download

# copy the rest of content
# COPY <src> <dst>
COPY . .

# declare arguments to catch from terminal command (docker build arguments)
ARG VERSION
ARG COMMIT
ARG BUILD_TIME

# $(head -n 1 go.mod | cut -d' ' -f2) <------get the module name from go.mod
# go build -ldflags <---- (Linker Flags) option allows to intercept string variables in source code right as it compiles and change their values without modifying the source code. 
# '-X ${MODULE_NAME}/internal/version.Version=${VERSION}' <-- -X string-inject VERSION value into version.Version
RUN MODULE_NAME=$(head -n 1 go.mod | cut -d' ' -f2) && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags " \
    -X '${MODULE_NAME}/internal/version.Version=${VERSION}' \
    -X '${MODULE_NAME}/internal/version.CommitHex=${COMMIT}' \
    -X '${MODULE_NAME}/internal/version.BuildTime=${BUILD_TIME}'" \
    -o app /app/cmd/web/main.go




# --------------------- RUN STAGE --------------------------- #

# Start fresh with pure Ubuntu 24.04
FROM ubuntu:24.04

# Install SSL certificates so your Go app can make secure outbound HTTPS calls
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# copy the compiled binary from build stage into current dir
COPY --from=builder /app/app .

# copy your SSR templates, static folders and migration folder from build stage
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# create empty dir for the board images
RUN mkdir -p /images/board

# create logs directory
RUN mkdir -p /logs

# the port to expose
EXPOSE 5000

# Run the app
CMD ["./app"]