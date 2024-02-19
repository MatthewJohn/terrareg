## Quick start

To easily get started, using a pre-built docker image:

    # Create secret key for session data
    export SECRET_KEY=$(python -c 'import secrets; print(secrets.token_hex())')

    # Run container, specifying secret key and admin password
    docker run -ti -p 5000:5000 -e PUBLIC_URL=http://localhost:5000 -e MIGRATE_DATABASE=True -e SECRET_KEY=$SECRET_KEY -e ADMIN_AUTHENTICATION_TOKEN=MySuperSecretPassword ghcr.io/matthewjohn/terrareg:latest


Visit http://localhost:5000 in your browser and follow the initial setup guide.
