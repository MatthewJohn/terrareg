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

# If private key has been set, inject into SSH directory
if [ "${SSH_PRIVATE_KEY}" != "" ]
then
    mkdir ~/.ssh
    chmod 700 ~/.ssh
    touch ~/.ssh/id_rsa
    chmod 600 ~/.ssh/id_rsa
    echo -e "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa
fi

# Run main executable
python ./terrareg.py
