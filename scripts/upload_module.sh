#!/bin/bash

base_url=$1

namespace=$2
name=$3
provider=$3
version=$4
file=$5

# Ensure source file exists
if ! test -f $file
then
   echo Source file does not exist: $file
   exit 1
fi

curl -X POST \
    "${base_url}/v1/${namespace}/${name}/${provider}/${version}/upload" \
    -H 'content-type: application/x-www-form-urlencoded' \
    --data-binary "@${file}"


