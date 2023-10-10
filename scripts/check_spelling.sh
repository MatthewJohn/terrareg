#!/bin/bash

set -x
set -e

pylint --disable all --enable spelling --spelling-dict=en_GB --spelling-private-dict-file=./scripts/dictionary.txt terrareg

