#!/bin/bash

set -e

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

echo Uploading module
curl -X POST \
    "${base_url}/v1/terrareg/modules/${namespace}/${name}/${provider}/${version}/upload" \
    -F file="@${file}"

echo Upload complete

echo 
echo "Would you like to 'publish' the module version? (Y/N)"
read publish

if [ "$publish" == "Y" ]
then
  curl -XPOST \
    "${base_url}/v1/terrareg/modules/${namespace}/${name}/${provider}/${version}/publish"
fi

