# single-stage: build and run in the same debian-based image
FROM golang:1.25.4

# install build deps and runtime libs
RUN apt-get update && apt-get install -y --no-install-recommends \
    git build-essential libsqlite3-dev ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /src

# cache modules
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# copy sources
COPY . .

# build with cgo enabled and place binaries into /usr/local/bin
RUN CGO_ENABLED=1 GOOS=linux go build -v -x -ldflags="-s -w" -o /usr/local/bin/sso-auth-server ./cmd/app
RUN CGO_ENABLED=1 GOOS=linux go build -v -x -ldflags="-s -w" -o /usr/local/bin/migrator ./cmd/migrator

# copy assets into image
RUN mkdir -p /build_assets
COPY config/appexample.yaml /build_assets/appexample.yaml
COPY migrations /build_assets/migrations

# place assets to runtime locations
RUN mkdir -p /app_config /migrations /data /var/log \
 && cp /build_assets/appexample.yaml /app_config/appexample.yaml \
 && cp -r /build_assets/migrations/* /migrations/ || true

# entrypoint
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

VOLUME ["/app_config", "/var/log", "/data"]

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD []
