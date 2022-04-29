#!/bin/bash

set -e
set -x

# Check if database upgrades are to be performed
if [ "${MIGRATE_DATABASE}" == "True" ]
then
    alembic upgrade head
fi

# Check whether to upgrade database and exit
if [ "${MIGRATE_DATABASE_ONLY}" == "True" ]
then
    exit
fi

# Run main executable
python ./terrareg.py
