#!/bin/bash

THIS_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $THIS_DIR

SUDO=`which sudo`
if [[ ! -z "$SUDO" ]]
then
  if ! sudo -n true 2>/dev/null; then 
    echo "Need sudo permissions to run webserver on port 80"
    sudo echo
  fi
fi

VANITY_HOST=`$GO run ../gen/cmd/gen/main.go vanity_host`
if [[ -z "$SUDO" ]]
then
  webroot=http://$VANITY_HOST $GO run ../cmd/server/main.go -port=80 &
else
  sudo -E env "PATH=$PATH" webroot=http://$VANITY_HOST $GO run ../cmd/server/main.go -port=80 &
fi

list_descendants () {
  local children=$(ps -o pid= --ppid "$1")
  for pid in $children
  do
    list_descendants "$pid"
  done
  echo "$children"
}

if [[ -z "$SUDO" ]]
then
  cp /etc/hosts /etc/hosts.bak
else
  sudo cp /etc/hosts /etc/hosts.bak
fi

if [[ -z "$SUDO" ]]
then
  trap 'mv /etc/hosts.bak /etc/hosts; kill $(list_descendants $$) >/dev/null 2>&1' EXIT
else
  trap 'sudo mv /etc/hosts.bak /etc/hosts; sudo kill $(list_descendants $$) >/dev/null 2>&1' EXIT
fi
until $(curl --output /dev/null --silent --head --fail http://localhost); do
    echo "waiting for webserver"
    sleep 1
done

hosts=`$GO run ../gen/cmd/gen/main.go hosts`
phony=`$GO run ../gen/cmd/gen/main.go phony`
export GOINSECURE=`$GO run ../gen/cmd/gen/main.go goinsecure`${GOINSECURE}

echo "temporarily remapping /etc/hosts"
if [[ -z "$SUDO" ]]
then
  hosts="$hosts" bash -c 'echo "$hosts" >> /etc/hosts'
else
  sudo hosts="$hosts" bash -c 'echo "$hosts" >> /etc/hosts'
fi

echo "creating phony go module"
cd ../doc/mod
if [[ ! -z "$DOCKER" ]]
then
  $GO clean -modcache
fi
rm go.* phony.go >/dev/null 2>&1
go mod init doc
echo "$phony" > phony.go

echo "downloading dependencies"
while IFS=';' read -ra MODS; do
  for i in "${MODS[@]}"; do
    repo=($(echo $i | awk -F= '{print $1}'))
    go get ${repo}@HEAD
  done
done <<< "$GO_MODULES"

go mod tidy
go mod download `go list -f '{{if not .Indirect}}{{.}}{{end}}' -m all | sed 1d | tr ' ' '@' | tr '\n' ' '`
