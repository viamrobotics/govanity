FROM golang:1.16.3 as builder
WORKDIR /app

ARG USE_GCP_SECRETS
ARG VANITY_HOST
RUN test -n "$VANITY_HOST"

# Copy local code to the container image.
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go
RUN go clean -modcache

# Set docs secrets
RUN docs_username=`cat etc/secrets/docs_username.txt` /bin/sh -c 'echo export DOCS_USERNAME=$docs_username >> /etc/new_environment'
RUN docs_password=`cat etc/secrets/docs_password.txt` /bin/sh -c 'echo export DOCS_PASSWORD=$docs_password >> /etc/new_environment'

# Download mods for docs
ENV GOPRIVATE=${VANITY_HOST}/*
RUN github_token=`cat etc/secrets/github_token.txt` /bin/sh -c 'git config --global url."https://${github_token}@github.com".insteadOf "https://github.com"'
RUN make docs_bootstrap
RUN rm -rf etc/secrets /root/.gitconfig

FROM alpine:3
WORKDIR /app
# Allow go binary to be run in alpine (different linkage)
RUN apk add --no-cache ca-certificates libc6-compat

# Copy sources/modules for docs
COPY --from=builder /usr/local/go/bin/go /usr/local/go/bin/go
COPY --from=builder /go/pkg/mod /root/go/pkg/mod
RUN mkdir -p /home/go/pkg && ln -s /root/go/pkg/mod /home/go/pkg/mod

COPY --from=builder /etc/new_environment /etc/environment
COPY --from=builder /app/server /app/server
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static
COPY --from=builder /app/etc/modules.txt /app/etc/modules.txt
COPY --from=builder /app/doc/mod /app/doc/mod

ENV GOMODPATH=/app/doc/mod
ENV PATH=/usr/local/go/bin:$PATH
ENV USE_GCP_SECRETS=${USE_GCP_SECRETS}
ENTRYPOINT /bin/sh -c "source /etc/environment && /app/server"
