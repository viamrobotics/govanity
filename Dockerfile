ARG GO_MODULES

FROM golang:1.16.3 as builder
WORKDIR /app
ARG GO_MODULES

RUN test -n "$GO_MODULES"

# Copy local code to the container image.
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o server_cmd cmd/server/main.go

FROM alpine:3
WORKDIR /app
ARG GO_MODULES

# Allow go binary to be run in alpine (different linkage)
RUN apk add --no-cache ca-certificates libc6-compat

COPY --from=builder /app/server_cmd /app/server_cmd

ENV GO_MODULES=$GO_MODULES
ENV PATH=/usr/local/go/bin:$PATH
RUN /app/server_cmd --verify
CMD ["/app/server_cmd"]
