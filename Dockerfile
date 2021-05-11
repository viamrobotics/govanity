ARG GO_MODULES

FROM golang:1.16.3 as builder
WORKDIR /app
ARG GO_MODULES

RUN test -n "$GO_MODULES"

# Copy local code to the container image.
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o server_cmd cmd/server/main.go

# Download mods for docs
RUN GITHUB_TOKEN=`cat etc/secrets/github_token.txt` /bin/sh -c 'git config --global url."https://${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"'
RUN GOPRIVATE=`go run gen/cmd/gen/main.go goinsecure` DOCKER=true make docs_bootstrap
RUN rm -rf etc/secrets /root/.gitconfig

FROM alpine:3
WORKDIR /app
ARG GO_MODULES

# Allow go binary to be run in alpine (different linkage)
RUN apk add --no-cache ca-certificates libc6-compat

# Copy sources/modules for docs
COPY --from=builder /usr/local/go/bin/go /usr/local/go/bin/go
COPY --from=builder /go/pkg/mod /root/go/pkg/mod
RUN mkdir -p /home/go/pkg && ln -s /root/go/pkg/mod /home/go/pkg/mod

COPY --from=builder /app/server_cmd /app/server_cmd
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/doc/mod /app/doc/mod

ENV GO_MODULES=$GO_MODULES
ENV PATH=/usr/local/go/bin:$PATH
CMD ["/app/server_cmd", "--enable-docs", "/app/doc/mod"]
