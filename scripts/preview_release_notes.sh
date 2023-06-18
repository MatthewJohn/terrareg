#!/bin/bash

docker run \
  -ti --rm -v `pwd`:/code \
  -w /code \
  --user=$UID \
  semantic-release \
  semantic-release \
    --dry-run --no-ci \
    --repository-url=file:///code \
     --branches=$(git rev-parse --abbrev-ref HEAD)
