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

if [[ -z "$SUDO" ]]
then
  USE_GCP_SECRETS="" webroot=http://$VANITY_HOST $GO run ../cmd/server/main.go -enable-docs=false -port=80 >/dev/null 2>&1 &
else
  sudo -E env "PATH=$PATH" USE_GCP_SECRETS="" webroot=http://$VANITY_HOST $GO run ../cmd/server/main.go -enable-docs=false -port=80 >/dev/null 2>&1 &
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
rm go.* phony.go >/dev/null 2>&1
go mod init doc
echo "$phony" > phony.go

echo "downloading dependencies"
while read repoMapping; do
	repo=($(echo $repoMapping | awk -F= '{print $1}'))
	go get ${repo}@HEAD
done < $THIS_DIR/modules.txt
go mod tidy
go mod download `go list -f '{{if not .Indirect}}{{.}}{{end}}' -m all | sed 1d | tr ' ' '@' | tr '\n' ' '`
