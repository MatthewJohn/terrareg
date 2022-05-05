FROM python:3.8

WORKDIR /

RUN apt-get update && apt-get install --assume-yes curl unzip git libxmlsec1 libxmlsec1-openssl libffi-dev && apt-get clean all

RUN wget https://github.com/terraform-docs/terraform-docs/releases/download/v0.16.0/terraform-docs-v0.16.0-linux-amd64.tar.gz && tar -zxvf terraform-docs-v0.16.0-linux-amd64.tar.gz && chmod +x terraform-docs && mv terraform-docs /usr/local/bin/ && rm terraform-docs-v0.16.0-linux-amd64.tar.gz

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .


ENTRYPOINT [ "bash", "scripts/entrypoint.sh" ]
