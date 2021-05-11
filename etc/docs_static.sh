#!/bin/bash

OUT_FILE="$1"
THIS_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $THIS_DIR

$GO run ../cmd/server/main.go --enable-docs --static-docs ../doc/mod >/dev/null 2>&1 &

# https://stackoverflow.com/a/26966800
kill_descendant_processes() {
    local pid="$1"
    local and_self="${2:-false}"
    if children="$(pgrep -P "$pid")"; then
        for child in $children; do
            kill_descendant_processes "$child" true
        done
    fi
    if [[ "$and_self" == true ]]; then
        kill -9 "$pid"
    fi
}

trap 'kill_descendant_processes $$ >/dev/null 2>&1' EXIT

until $(curl --output /dev/null --silent --head --fail http://localhost:8080); do
    echo "waiting for webserver"
    sleep 1
done

echo "generating docs"
rm -rf dist
wget \
    --quiet \
    --recursive \
    --no-verbose \
    --convert-links \
    --page-requisites \
    --adjust-extension \
    --content-disposition \
    --execute=robots=off \
    --directory-prefix="dist" \
    --no-host-directories \
    --domains localhost \
    --exclude-domains pkg.go.dev \
    http://localhost:8080

find dist -type f -print0 | xargs -0 sed -i "s/http:\/\/localhost:8080/https:\/\/pkg.go.dev/g"
tar -czf $OUT_FILE dist
rm -rf dist
