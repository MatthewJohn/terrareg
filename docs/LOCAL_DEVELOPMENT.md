# Local Development

## Running with docker-compose for Development

A docker-compose file is avaiable to simplify launching terrareg for local testing and development. This will let you run terrareg with an SSL certificate, allowing terraform cli to access modules while developing or testing the software. In addition, the root folder is mounted in the container allowing for rapid development and testing without rebuilding the container.

Using docker-compose will spin up a stack of containers including:

  * Traefik
  * docker-socket-proxy
  * terrarreg
  * mysql
  * phpmyadmin
  * minio (S3 storage)

__*NOTE: Traefik requires exposing the docker socket to thhe container. Please see [here](https://doc.traefik.io/traefik/providers/docker/#docker-api-access) for more information. This implementation utilizes [docker-socket-proxy](https://github.com/Tecnativa/docker-socket-proxy) to limit the exposure*__

### Install mkcert

mkcert is used to create a local CA for generating self signed SSL Certificates that are automatically trusted by your local system. If you wish to manually generate the SSL Certificates and add them to your system and browser trust stores you can skip this.

mkcert can be installed on [Linux](https://github.com/FiloSottile/mkcert#linux), [MacOS](https://github.com/FiloSottile/mkcert#macos), [Windows](https://github.com/FiloSottile/mkcert#windows) & WSL (See notes below for WSL). After installing run the following command to create a new Local CA:

    mkcert -install

#### WSL Setup

If you are using WSL2, install mkcert for Windows Then run the following in powershell:

    setx CAROOT "$(mkcert -CAROOT)"; If ($Env:WSLENV -notlike "*CAROOT/up:*") { setx WSLENV "CAROOT/up:$Env:WSLENV" }

This will set WSL2 to use the Certificate Authority File in Windows.

Now install mkcert for Linux inside of your WSL2 instance. Once you've done this run the following command to see that mkcert is referencing the path on your C drive:
     mkcert -CAROOT

After confirming the CAROOT path maps to your windows user (should look like /mnt/c/Users/YourUser/AppData/Local/mkcert) the CA Cert needs to be installed inside of WSL so that terraform recognizes your Local CA as Trusted. This can be done by running:

    mkcert -install

Once this has been completed all remaining commands should be run inside of WSL.

### Generate Local Development SSL Certs

Now that mkcert is installed and a Local CA has been generated it's time to generate an SSL Certificate for Traefik to use when proxying to the terrareg container. To do this run:

    mkdir -p certs
    mkcert -cert-file certs/local-cert.pem -key-file certs/local-key.pem "app.localhost" "*.app.localhost" 

### Container .env files

You will find an EXAMPLE.env file that is used to configure the stack. Copy this to .env and adjust the configuration options as documented below. The key/value pairs in thie file are passed as Environment variables to the terrareg container.

Make sure to change the following variables in the .env file before launching:

* SECRET_KEY
* ADMIN_AUTHENTICATION_TOKEN

If you wish to mount a folder containing your ssh keys into the container see the Volumes section for terrareg in _docker-compose.yml_ for an example.

### Run the Stack

Once mkcert has been installed & configured with a local CA and SSL Certificates it's time to start up the stack.

    docker-compose up -d

Wait a moment for everything to come online. Terrareg will become available after MySQL comes online.

You can access the stack at the following URLs:

  * terrareg - https://terrareg.app.localhost/
  * phpmyadmin - https://phpmyadmin.app.localhost/
  * traefik - https://traefik.app.localhost

Because everything referencing localhost routes to 172.0.0.1 no special host file entries are required.

## Building locally and running

```
# Clone the repository
git clone https://github.com/matthewJohn/terrareg
cd terrareg

# Optionally create a virtualenv
virtualenv -ppython3 venv
. venv/bin/activate

# Install libmagic
## For OS X:
brew install libmagic

## For Ubuntu
sudo apt-get install libmagic1

# Install depdencies:
pip install poetry
poetry install --no-root --with=dev

# Initialise database and start server:
poetry run alembic upgrade head

# Set random admin authentication token - the password used for authenticating as the built-in admin user
export ADMIN_AUTHENTICATION_TOKEN=MySuperSecretPassword
# Set random secret key, used encrypting client session data
export SECRET_KEY=$(python -c 'import secrets; print(secrets.token_hex())')

# Obtain terraform-docs, Tfsec and Infracost
mkdir bin
export PATH=$PATH:`pwd`/bin
if [ "$(uname -m)" == "aarch64" ]; then arch=arm64; else arch=amd64; fi
wget https://github.com/terraform-docs/terraform-docs/releases/download/v0.16.0/terraform-docs-v0.16.0-linux-${arch}.tar.gz && tar -zxvf terraform-docs-v0.16.0-linux-${arch}.tar.gz terraform-docs && chmod +x terraform-docs && mv terraform-docs ./bin/ && rm terraform-docs-v0.16.0-linux-${arch}.tar.gz
wget https://github.com/aquasecurity/tfsec/releases/download/v1.26.0/tfsec-linux-${arch} -O ./bin/tfsec && chmod +x ./bin/tfsec
wget https://github.com/infracost/infracost/releases/download/v0.10.10/infracost-linux-${arch}.tar.gz && tar -zxvf infracost-linux-${arch}.tar.gz infracost-linux-${arch} && mv infracost-linux-${arch} ./bin/infracost && chmod +x ./bin/infracost && rm infracost-linux-${arch}.tar.gz

# Run the server
poetry run python ./terrareg.py
```

The site can be accessed at http://localhost:5000

## Generating DB changes

Once changes are made to a

```
# Ensure database is up-to-date before generating schema migrations
poetry run alembic upgrade head

# Generate migration
poetry run alembic revision --autogenerate
```

## Applying DB changes

```
poetry run alembic upgrade head
```

## Running tests

```
# Install dev requirements
pip install poetry
poetry install --no-root --with=dev

# Run all tests
poetry run pytest

# Running unit/integration/selenium tests individually
poetry run pytest ./test/unit
poetry run pytest ./test/integration
poetry run pytest ./test/selenium

# Running a specific test
poetry run pytest -k test_setup_page
```
