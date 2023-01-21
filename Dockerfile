FROM python:3.10

WORKDIR /

RUN apt-get update && \
    apt-get install --assume-yes \
        curl unzip git \
        pkg-config libxml2-dev libxmlsec1-dev libxmlsec1-openssl xmlsec1 libgraphviz-dev && \
    apt-get clean all

RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then \
      arch=arm64; \
    else \
      arch=amd64; \
    fi; \
    wget https://github.com/terraform-docs/terraform-docs/releases/download/v0.16.0/terraform-docs-v0.16.0-linux-${arch}.tar.gz && tar -zxvf terraform-docs-v0.16.0-linux-${arch}.tar.gz && chmod +x terraform-docs && mv terraform-docs /usr/local/bin/ && rm terraform-docs-v0.16.0-linux-${arch}.tar.gz'

RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then \
      arch=arm64; \
    else \
      arch=amd64; \
    fi; \
    wget https://github.com/aquasecurity/tfsec/releases/download/v1.26.0/tfsec-linux-${arch} -O /usr/local/bin/tfsec && \
    chmod +x /usr/local/bin/tfsec'

# Download infracost
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then \
      arch=arm64; \
    else \
      arch=amd64; \
    fi; \
    wget https://github.com/infracost/infracost/releases/download/v0.10.10/infracost-linux-${arch}.tar.gz -O /tmp/infracost.tar.gz && \
    tar -zxvf /tmp/infracost.tar.gz infracost-linux-${arch} && \
    mv infracost-linux-${arch} /usr/local/bin/infracost && \
    chmod +x /usr/local/bin/infracost && \
    rm /tmp/infracost.tar.gz'

# Download tfswitch
RUN bash -c 'curl -L https://raw.githubusercontent.com/warrensbox/terraform-switcher/release/install.sh | bash'

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY LICENSE .
COPY alembic.ini .
COPY terrareg.py .
COPY terrareg terrareg
COPY scripts scripts

ENV MANAGE_TERRAFORM_RC_FILE=True

EXPOSE 5000

ENTRYPOINT [ "bash", "scripts/entrypoint.sh" ]
