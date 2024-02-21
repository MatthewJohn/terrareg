# Deployment

## Docker environment variables

The following environment variables are available to configure the docker container

### MIGRATE_DATABASE

Whether to perform a database migration on container startup.

Set to `True` to enable database migration

*Note:* Be very careful when scaling the application. There should never be more than one instance of Terrareg running with `MIGRATE_DATABASE` set to `True` during an upgrade.

When upgrading, scale the application to a single instance before upgrading to a newer version.

Alternatively, set `MIGRATE_DATABASE` to `False` and run a dedicated instance for performing database upgrades.
Use `MIGRATE_DATABASE_ONLY` to run an instance that will perform the necessary database migrations and immediately exit.

Default: `False`

### MIGRATE_DATABASE_ONLY

Whether to perform database migration and exit immediately.

This is useful for scheduling migrations by starting a 'migration' instance of the application.

Set to `True` to exit after migration.

The `MIGRATE_DATABASE` environment variable must also be set to `True` to perform the migration, otherwise nothing will be performed and the container will exit.

Default: `False`

### SSH_PRIVATE_KEY

Provide the contents of the SSH key to perform git clones.

This is an alternative to mounting the '.ssh' directory of the root user.

Default: ''


## Application environment variables

For a list of available application configuration environment variables, please see [doc/CONFIG](./CONFIG.md)


## Database Migrations

Terrareg can be deployed via Docker and scaled out to support high-availability and load requirements.

However, whilst perform upgrades with database migrations, it's important to ensure that only one container performs the database migration step.

This can be accomplished in two ways:

 * Scale down to 1 container when performing upgrades that container a database migration
 * Run a single dedicated container that performs the database upgrades.

In either situation, when performing a database upgrade, it is highly recommended that any containers serving the web-application are stopped.

### On-the-fly DB migrations

To enable database migrations in all containers (assuming the service will be scaled to a single container during upgrade), set the environment variable `MIGRATE_DATABASE` to `True`.

### Dedicated DB migration container

To dedicate a single container to DB migrations, set `MIGRATE_DATABASE` to `False` on all containers running the web application and create a new container

## Allowing Terrareg to Communicate with itself

During module extraction/analysis, Terrareg will need to communicate with itself, which is required during cost analysis and graph generation.

To configure this, set the [DOMAIN_NAME](./CONFIG.md#domain_name) configuration.

To ensure terraform does not generate unecessary analytics in the module, terrareg must manage the .terraformrc file in the user's home directory.
This functionality is enabled by default in the docker container, by disabled outside of it. To enable/disable this functionality, see [MANAGE_TERRAFORM_RC_FILE](./CONFIG.md#manage_terraform_rc_file)

## Docker storage

### Module/Provider data

If [module hosting](./modules/storage.md) is being used, ensure that a directory is mounted into the container for storing module data.
This path can be customised by setting [DATA_DIRECTORY](./CONFIG.md#data_directory)

Alternative, an S3 bucket may be used for storage. Configure the [DATA_DIRECTORY](./CONFIG.md#data_directory) with an `s3://` URL, as per the example. When using S3, the container must have native access to an s3 bucket, that is, either on an EC2 instance with S3 permissions, an ECS container with an IAM role that has permission to the s3 bucket or have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables configured.

If S3 is used for the data directory, the [UPLOAD_DIRECTORY](./CONFIG.md#upload_directory) must be configured to a path, as this usually default to the DATA_DIRECTORY but does not support S3.

## Database URL

It is recommended to use an external database when using in production.
Terrareg has been tested with MariaDB 10.X and is incompattible with MySQL <= 8.x

A database URL should be configured, otherwise Terrareg will default to a local sqlite database (though this _could_ be mounted via a docker volume, it certainly shouldn't be used for multiple containers).

This should be configured by setting [DATABASE_URL](./CONFIG.md#database_url)

## Listen port

The port that Terrareg listens on can be configured with [LISTEN_PORT](./CONFIG.md#listen_port)

## SSL

Although Terrareg can be deployed without SSL - this is only recommended for testing and local development.
Aside from the usualy reasons for using SSL, it is also required for Terraform to communicate with the registry to obtain modules as a registry provider. If SSL is not used, Terrareg will fall-back to providing Terrareform examples using a 'http' download URL for Terraform.

Terrareg must be configured with the URL that the registry is accessible. To configure this, please see [PUBLIC_URL](./CONFIG.md#public_url)


### Enabling SSL on the application

SSL can be enabled on Terrareg itself - the certificates must be mounted inside the container (or be available on the filesystem, if running outside of a docker container) and the absolute path can be provided using the environment variables [SSL_CERT_PRIVATE_KEY](./CONFIG.md#ssl_cert_private_key) and [SSL_CERT_PUBLIC_KEY](./CONFIG.md#ssl_cert_public_key).

If Terrareg is being run outside of a docker container, these can be provided as command line arguments `--ssl-cert-private-key` and `--ssl-cert-public-key`.

### Offloading SSL using a reverse proxy

SSL can also be provided by a reverse proxy in front of Terrareg and traffic to the Terrareg container can be served via HTTP.
