#!/bin/bash

OUT_FILE="$1"
THIS_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $THIS_DIR

$GO run ../cmd/server/main.go --static-docs >/dev/null 2>&1 &

list_descendants () {
  local children=$(ps -o pid= --ppid "$1")
  for pid in $children
  do
    list_descendants "$pid"
  done
  echo "$children"
}

trap 'kill $(list_descendants $$) >/dev/null 2>&1' EXIT

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
