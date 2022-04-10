# terrareg

Terraform Registry

## Getting started

Install depdencies:

    pip install -r requirements.txt

Start server:

    python ./terrareg.py


Upload a terraform module:

    terrareg_root=$PWD
    
    cd source/of/my/module
    
    # OPTIONAL: Create an terrareg meta-data file
    echo '{ "description": "My first module", "owner": "ME!", "source": "https://github.com/me/my-tf-module" }' > ./terrareg.json
    
    # Zip up module
    zip -r * ../my-tf-module.zip
    
    # Upload to terrareg
    bash $terrareg_root/scripts/upload_module.sh http://localhost:5000 me my-tf-module aws 1.0.0 source/of/my/my-tf-module.zip


Navigate to http://localhost:5000 to get started, or http://localhost/modules/me/my-tf-module to see the uploaded example!

## Local development

Since terraform requires HTTPS with a valid SSL cert, this must be provided in local development

### Without a valid SSL cert:

    docker run --rm -it \
      -e "UPSTREAM_DOMAIN=local-dev" \
      --add-host=laptop-ext:<local IP> \
      -e "UPSTREAM_PORT=5000" \
      -e "PROXY_DOMAIN=localdev.example.com" \
      -l "com.dnsdock.name=proxy" \
      -p 443:443 \
      -v $(pwd)/ca:/etc/nginx/ca \
      outrigger/https-proxy:1.0
    
    # Add CA cert to local trusted certificates (I know! :( )
    sudo cp ca/rootCA.crt /usr/local/share/ca-certificates/
    sudo update-ca-certificates

@TODO Move running terraform in local development to a docker container


