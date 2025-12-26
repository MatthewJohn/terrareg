# Multi-stage build for Terrareg Go implementation
FROM golang:1.25.4-trixie AS builder

ARG VERSION
ARG http_proxy
ARG https_proxy
ARG HTTP_PROXY
ARG HTTPS_PROXY

WORKDIR /app

# Install build dependencies
RUN apt-get update && \
    apt-get install --assume-yes \
        pkg-config libxml2-dev libxmlsec1-dev libxmlsec1-openssl \
        xmlsec1 libgraphviz-dev libmagic1 && \
    apt-get clean all

# Copy Go modules files
COPY terrareg-go/go.mod terrareg-go/go.sum ./

# Download dependencies
RUN http_proxy=$http_proxy https_proxy=$https_proxy \
    go mod download

# Copy source code
COPY terrareg-go/ ./

# Build the Go binary
RUN http_proxy=$http_proxy https_proxy=$https_proxy \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=$VERSION" \
    -o server ./cmd/server/main.go


# Final stage - minimal runtime image
FROM debian:bookworm-slim

ARG VERSION

# Install runtime dependencies
RUN apt-get update && \
    apt-get install --assume-yes \
        curl unzip git wget zip \
        ca-certificates libxmlsec1-openssl xmlsec1 \
        libgraphviz-dev libmagic1 && \
    apt-get clean all

# Download terraform-docs
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then arch=arm64; else arch=amd64; fi; \
    wget https://github.com/terraform-docs/terraform-docs/releases/download/v0.19.0/terraform-docs-v0.19.0-linux-${arch}.tar.gz && \
    tar -zxvf terraform-docs-v0.19.0-linux-${arch}.tar.gz && \
    chmod +x terraform-docs && mv terraform-docs /usr/local/bin/ && \
    rm terraform-docs-v0.19.0-linux-${arch}.tar.gz'

# Download tfsec
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then arch=arm64; else arch=amd64; fi; \
    wget https://github.com/aquasecurity/tfsec/releases/download/v1.28.4/tfsec-linux-${arch} -O /usr/local/bin/tfsec && \
    chmod +x /usr/local/bin/tfsec'

# Download Infracost
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then arch=arm64; else arch=amd64; fi; \
    wget https://github.com/infracost/infracost/releases/download/v0.10.22/infracost-linux-${arch}.tar.gz -O /tmp/infracost.tar.gz && \
    tar -zxvf /tmp/infracost.tar.gz infracost-linux-${arch} && \
    mv infracost-linux-${arch} /usr/local/bin/infracost && \
    chmod +x /usr/local/bin/infracost && \
    rm /tmp/infracost.tar.gz'

# Download tfswitch
RUN bash -c 'curl -L https://raw.githubusercontent.com/warrensbox/terraform-switcher/master/install.sh | bash /dev/stdin 1.2.2'

# Download tfplugindocs
RUN bash -c 'if [ "$(uname -m)" == "aarch64" ]; \
    then arch=arm64; else arch=amd64; fi; \
    wget https://github.com/hashicorp/terraform-plugin-docs/releases/download/v0.16.0/tfplugindocs_0.16.0_linux_${arch}.zip -O /tmp/tfplugindocs.zip && \
    unzip /tmp/tfplugindocs.zip tfplugindocs && \
    mv tfplugindocs /usr/local/bin/ && \
    chmod +x /usr/local/bin/tfplugindocs && \
    rm /tmp/tfplugindocs.zip'

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/server /app/terrareg-go/server

# Copy static assets, templates, and scripts from Python implementation
COPY terrareg/ ./terrareg/
COPY scripts ./scripts
COPY alembic.ini .
COPY LICENSE .
COPY LICENSE.third-party .

# Create licenses directory
RUN mkdir -p licenses

# Download licenses for installed binaries
RUN mkdir -p licenses/terraform-docs && \
    wget https://github.com/terraform-docs/terraform-docs/raw/master/LICENSE -O ./licenses/terraform-docs/LICENSE

RUN mkdir -p licenses/tfsec && \
    wget https://github.com/aquasecurity/tfsec/raw/master/LICENSE -O ./licenses/tfsec/LICENSE

RUN mkdir -p licenses/infracost && \
    wget https://github.com/infracost/infracost/raw/master/LICENSE -O ./licenses/infracost/LICENSE

RUN mkdir -p licenses/terraform-switcher && \
    wget https://github.com/warrensbox/terraform-switcher/raw/master/LICENSE -O ./licenses/terraform-switcher/LICENSE

RUN mkdir -p licenses/go && \
    wget https://github.com/golang/go/raw/master/LICENSE -O ./licenses/go/LICENSE

RUN mkdir -p licenses/tfplugindocs && \
    wget https://github.com/hashicorp/terraform-plugin-docs/raw/main/LICENSE -O ./licenses/tfplugindocs/LICENSE

# Copy licenses for JS/CSS
RUN mkdir -p licenses/static && \
    bash -c 'for n in js css; do \
        if [ -d "/app/terrareg/static/$n" ]; then \
            pushd /app/terrareg/static/$n; \
            for i in *; do \
                if [ -d "$i" ] && [ -f "$i/LICENSE" ]; then \
                    mkdir -p /app/licenses/static/$i; \
                    cp $i/LICENSE /app/licenses/static/$i/; \
                fi; \
            done; \
            popd; \
        fi; \
    done'

# Set version
RUN echo "$VERSION" > /app/terrareg/version.txt

ENV MANAGE_TERRAFORM_RC_FILE=True

EXPOSE 5000

# Use the Go server binary
CMD ["/app/terrareg-go/server"]
