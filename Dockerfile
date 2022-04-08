FROM python:3

WORKDIR /

RUN apt-get update && apt-get install --assume-yes curl unzip && apt-get clean all
RUN curl https://github.com/terraform-docs/terraform-docs/releases/download/v0.16.0/terraform-docs-v0.16.0-linux-amd64.tar.gz -o

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY . .

ENTRYPOINT [ "python", "terrareg.py" ]

