# Build container
FROM golang AS Build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY cmd ./cmd
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -o /gpodder2go

FROM alpine
RUN mkdir /data
WORKDIR /data
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /data /entrypoint.sh
COPY --from=Build /gpodder2go /gpodder2go

# Set label
LABEL org.opencontainers.image.title="GPodder2Go Docker Image" \
      org.opencontainers.image.description="gpodder2go is a self-hosted server to handle podcast subscriptions management for gpodder clients" \
      org.opencontainers.image.documentation="https://github.com/dontobi/gpodder2go#readme" \
      org.opencontainers.image.authors="Tobias Schug <github@myhome.zone>" \
      org.opencontainers.image.url="https://github.com/dontobi/gpodder2go" \
      org.opencontainers.image.source="https://github.com/dontobi/gpodder2go" \
      org.opencontainers.image.base.name="docker.io/library/alpine:latest" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${DATI}"

# Set Variables
ENV NO_AUTH=false

# Set Ports
EXPOSE 3005

# Set Volumes
VOLUME /data

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]
