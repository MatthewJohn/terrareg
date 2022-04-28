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

## Terrareg module metadata

A metadata file can be provided each an uploaded module's archive to provide additional metadata to terrareg.

For an example, please see: [docs/example-terrareg-module-metadata.json](./docs/example-terrareg-module-metadata.json)

The following attributes are available at the root of the JSON object:

|Key |Description|
--- | --- |
|owner|Name of the owner of the module|
|source|Link to the original source code (used for links on the module page)|
|description|Description of the module|
|variable_template|Structure holding required input variables for module, used for 'Usage Builder'. See table below|
|artifact_location|Optional variable to use an external URL as source of terraform module. Add `{module_version}` as placeholder for the module version. If not provided, the module will be hosted by terrareg.|

### Usage builder configuration

The usage builder requires an array of objects, which define the name, type and description of the variable.

In the following the 'config input' refers to the HTML inputs that provide the user with the ability to select/enter values. The 'terraform input' refers to the value used for the variable in the outputted terraform example.

There are common attributes that can be added to each of variable objects, which include:

|Attribute |Description |Default|
--- | --- | ---|
|name|The name of the 'config input'. This is also used as the module variable in the 'terraform input'.|Required|
|type|The type of the input variable, see table below.|Required|
|quote_value|Boolean flag to determine whether the value generated is quoted for the 'terraform input'.|false|
|additional_help|A description that is provided, along with the terraform variable description in the usage builder|Empty|



|Variable type|Description|Type specific attributes|
--- | --- | ---|
|text|A plain input text box for users to provide a value that it directly used as the 'terraform input'||
|boolean|Provides a checkbox that results in a true/false value as the 'terraform input'||
|static|This does not appear in the 'Usage Builder' 'config input' table, but provides a static value in the 'terraform input'||
|select|Provides a dropdown for the user to select from a list of choices|"choices" must be added to the object, which may either be a list of strings, or a list of objects. If using a list of objects, a "name" and "value" must be provided. Optionally an "additional_content" attribute can be added to the choice, which provides additional terraform to be added to the top of the terraform example. The main variable object may also contain a "allow_custom" boolean attribute, which allows the user to enter a custom text input.|


## Local development

Since terraform requires HTTPS with a valid SSL cert, this must be provided in local development

On linux, by default, non-privileged users cannot listen on privileged ports, so the following can be used to route requests locally to port 5000:

```
sudo iptables -t nat -I OUTPUT -p tcp -d 127.0.0.1 --dport 443 -j REDIRECT --to-ports 5000
```

Example to run in local development environment:
```
virtualenv -ppython3.8 venv
. venv/bin/activate
pip install -r requirements.txt
pip install -r requirements-dev.txt

# Without SSL cert
ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER=False ALLOW_CUSTOM_GIT_URL_MODULE_VERSION=False GIT_PROVIDER_CONFIG='[{"name": "Github", "base_url": "https://github.com/{namespace}/{module}", "clone_url": "ssh://git@github.com:{namespace}/{module}.git", "browse_url": "https://github.com/{namespace}/{module}/tree/{tag}/{path}"}, {"name": "Bitbucket", "base_url": "https://bitbucket.org/{namespace}/{module}", "clone_url": "ssh://git@bitbucket.org:{namespace}/{module}-{provider}.git", "browse_url": "https://bitbucket.org/{namespace}/{module}-{provider}/src/{tag}/{path}"}, {"name": "Gitlab", "base_url": "https://gitlab.com/{namespace}/{module}", "clone_url": "ssh://git@gitlab.com:{namespace}/{module}-{provider}.git", "browse_url": "https://gitlab.com/{namespace}/{module}-{provider}/-/tree/{tag}/{path}"}]' SECRET_KEY=ec9b8cc5ed0404acb3983b7836844d828728c22c28ecbed9095edef9b7489e85 ADMIN_AUTHENTICATION_TOKEN=password ANALYTICS_AUTH_KEYS=xxxxxx.atlasv1.zzzzzzzzzzzzz:dev,xxxxxx.atlasv1.xxxxxxxxxx:prod VERIFIED_MODULE_NAMESPACES=hashicorp TRUSTED_NAMESPACES=test DEBUG=True AUTO_PUBLISH_MODULE_VERSIONS=False LISTEN_PORT=5001 python ./terrareg.py

# With SSL Cert
# Add the following argument
#  --ssl-cert-private-key ./example/ssl-certs/private.pem --ssl-cert-public-key ./example/ssl-certs/public.pem

```

