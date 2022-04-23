#!/bin/bash

base_url=$1

namespace=$2
name=$3
provider=$4
version=$5
file=$6

# Ensure source file exists
if ! test -f $file
then
   echo Source file does not exist: $file
   exit 1
fi

curl -X POST \
    "${base_url}/v1/terrareg/modules/${namespace}/${name}/${provider}/${version}/upload" \
    -F file="@${file}" -vvvv


