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
curl -k -X POST \
    "${base_url}/v1/terrareg/modules/${namespace}/${name}/${provider}/${version}/upload" \
    -F file="@${file}" -H "X-Terrareg-ApiKey: $UPLOAD_API_KEY"

echo Upload complete

echo 
echo "Would you like to 'publish' the module version? (Y/N)"
read publish

if [ "$publish" == "Y" ]
then
  curl -k -XPOST -H "X-Terrareg-ApiKey: $PUBLISH_API_KEY" \
    "${base_url}/v1/terrareg/modules/${namespace}/${name}/${provider}/${version}/publish"
fi

