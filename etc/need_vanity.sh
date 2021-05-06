#!/bin/bash

THIS_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
cd $THIS_DIR

hosts=`$GO run ../gen/cmd/gen/main.go hosts`
if [[ ! -z $hosts && ! -z $VANITY_HOST_NOT_SET  ]]
then
  echo "Need to set VANITY_HOST if hosting vanity modules"
  exit 1
fi
