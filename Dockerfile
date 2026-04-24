FROM python:3.13-slim

ARG VERSION

# renovate: datasource=github-releases depName=terraform-docs/terraform-docs
ARG TERRAFORM_DOCS_VERSION=v0.21.0
# renovate: datasource=github-releases depName=aquasecurity/tfsec
ARG TFSEC_VERSION=v1.28.14
# renovate: datasource=github-releases depName=infracost/infracost
ARG INFRACOST_VERSION=v0.10.43
# renovate: datasource=github-releases depName=warrensbox/terraform-switcher
ARG TERRAFORM_SWITCHER_VERSION=1.17.1
# renovate: datasource=github-releases depName=hashicorp/terraform-plugin-docs
ARG TFPLUGINDOCS_VERSION=0.24.0
# renovate: datasource=golang-version depName=golang
ARG GO_VERSION=1.25.7

WORKDIR /

RUN apt-get update && \
  apt-get install --assume-yes \
  curl unzip git wget zip git \
  pkg-config libxml2-dev libxmlsec1-dev libxmlsec1-openssl xmlsec1 libgraphviz-dev libmagic1 \
  gcc g++ libffi-dev python3-gpg zlib1g-dev && \
  apt-get clean all

RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
  then \
  arch=arm64; \
  else \
  arch=amd64; \
  fi; \
  wget https://github.com/terraform-docs/terraform-docs/releases/download/${TERRAFORM_DOCS_VERSION}/terraform-docs-${TERRAFORM_DOCS_VERSION}-linux-${arch}.tar.gz && tar -zxvf terraform-docs-${TERRAFORM_DOCS_VERSION}-linux-${arch}.tar.gz && chmod +x terraform-docs && mv terraform-docs /usr/local/bin/ && rm terraform-docs-${TERRAFORM_DOCS_VERSION}-linux-${arch}.tar.gz'

RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
  then \
  arch=arm64; \
  else \
  arch=amd64; \
  fi; \
  wget https://github.com/aquasecurity/tfsec/releases/download/${TFSEC_VERSION}/tfsec-linux-${arch} -O /usr/local/bin/tfsec && \
  chmod +x /usr/local/bin/tfsec'

# Download Infracost
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then \
      arch=arm64; \
    else \
      arch=amd64; \
    fi; \
    wget https://github.com/infracost/infracost/releases/download/${INFRACOST_VERSION}/infracost-linux-${arch}.tar.gz -O /tmp/infracost.tar.gz && \
    tar -zxvf /tmp/infracost.tar.gz infracost-linux-${arch} && \
    mv infracost-linux-${arch} /usr/local/bin/infracost && \
    chmod +x /usr/local/bin/infracost && \
    rm /tmp/infracost.tar.gz'

# Download tfswitch
RUN bash -c 'curl -L https://raw.githubusercontent.com/warrensbox/terraform-switcher/master/install.sh | bash /dev/stdin ${TERRAFORM_SWITCHER_VERSION}'

# Install go
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
  then \
  arch=arm64; \
  else \
  arch=amd64; \
  fi; \
  wget https://go.dev/dl/go${GO_VERSION}.linux-${arch}.tar.gz -O /tmp/go.tar.gz && \
  tar -zxvf /tmp/go.tar.gz -C /usr/local && \
  rm /tmp/go.tar.gz'
ENV PATH=$PATH:/usr/local/go/bin

# Install github.com/hashicorp/terraform-plugin-docs
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
  then \
  arch=arm64; \
  else \
  arch=amd64; \
  fi; \
  wget https://github.com/hashicorp/terraform-plugin-docs/releases/download/v${TFPLUGINDOCS_VERSION}/tfplugindocs_${TFPLUGINDOCS_VERSION}_linux_${arch}.zip -O /tmp/tfplugindocs.zip && \
  unzip /tmp/tfplugindocs.zip tfplugindocs && \
  mv tfplugindocs /usr/local/bin/ && \
  chmod +x /usr/local/bin/tfplugindocs && \
  rm /tmp/tfplugindocs.zip'

WORKDIR /app

COPY pyproject.toml poetry.lock .
ARG PYPI_PROXY
RUN if test ! -z "$PYPI_PROXY"; then pip_args="--index=$PYPI_PROXY --trusted-host=$(echo $PYPI_PROXY | sed 's#https*://##g' | sed 's#/.*##g')"; else pip_args=""; fi; \
  http_proxy= https_proxy="" pip install poetry $pip_args
RUN poetry config virtualenvs.in-project true
# Build lxml and xmlsec from source to match system libraries
RUN poetry config installer.no-binary lxml,xmlsec

RUN if test ! -z "$PYPI_PROXY"; then \
  poetry source add --priority=primary packages $PYPI_PROXY; \
  http_proxy= https_proxy= poetry lock; \
  fi
ARG POETRY_INSTALLER_MAX_WORKERS=4
RUN https_proxy= http_proxy= poetry install --no-root

RUN mkdir bin licenses

# Create licenses for python packages
RUN if test ! -z "$PYPI_PROXY"; then pip_args="--index=$PYPI_PROXY --trusted-host=$(echo $PYPI_PROXY | sed 's#https*://##g' | sed 's#/.*##g')"; else pip_args=""; fi; \
  http_proxy= https_proxy= pip install pip-licenses $pip_args && \
  pip-licenses --with-system --with-license-file --format=plain-vertical > licenses/LICENSES.python && \
  pip uninstall --yes pip-licenses
# Copy licenses for deb packages
RUN mkdir licenses/deb
RUN bash -c 'pushd /usr/share/doc; for i in *; do mkdir /app/licenses/deb/$i; cp $i/{LICENSE,NOTICE,copyright} /app/licenses/deb/$i/; done; rmdir /app/licenses/deb/*; popd'
# Get licenses for installed binaries
RUN mkdir licenses/terraform-docs && wget https://github.com/terraform-docs/terraform-docs/raw/master/LICENSE -O ./licenses/terraform-docs/LICENSE
RUN mkdir licenses/tfsec && wget https://github.com/aquasecurity/tfsec/raw/master/LICENSE -O ./licenses/tfsec/LICENSE
RUN mkdir licenses/infracost && wget https://github.com/infracost/infracost/raw/master/LICENSE -O ./licenses/infracost/LICENSE
RUN mkdir licenses/terraform-switcher && wget https://github.com/warrensbox/terraform-switcher/raw/master/LICENSE -O ./licenses/terraform-switcher/LICENSE
RUN mkdir licenses/go && wget https://github.com/golang/go/raw/master/LICENSE -O ./licenses/go/LICENSE
RUN mkdir licenses/tfplugindocs && wget https://github.com/hashicorp/terraform-plugin-docs/raw/main/LICENSE -O ./licenses/tfplugindocs/LICENSE

COPY LICENSE .
COPY LICENSE.third-party .
COPY alembic.ini .
COPY terrareg.py .
COPY terrareg terrareg
COPY scripts scripts
RUN echo "$VERSION" > terrareg/version.txt

# Copy licenses for JS/CSS
RUN mkdir licenses/static
RUN bash -c 'for n in js css; do pushd /app/terrareg/static/$n; for i in *; do if [ -d $i ]; then mkdir /app/licenses/static/$i; cp $i/LICENSE /app/licenses/static/$i/; fi; done; popd; done'

ENV MANAGE_TERRAFORM_RC_FILE=True

EXPOSE 5000

CMD [ "bash", "scripts/entrypoint.sh" ]
