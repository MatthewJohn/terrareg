# Single Sign-on

*NOTE* OpenID and SAML2 authentication is currently experimental.

It is advised to use with caution and avoid using on publicly hosted/accessible instances.

### OpenID Connect

To configure OpenID connect, setup an application in an identity provider (IdP) with the following:

| | |
| --- | --- |
| Grant type | Authorization Code |
| Sign-in redirect URI | `https://<terrareg-instance-domain>/openid/callback` |
| Sign-out URI | `https://<terrareg-instance-domain>/logout` |
| Initiate login URI | `https://<terrareg-instance-domain>/openid/login` |
| Login flow | Redirect to app to initiate login |

Obtain the client ID, client secret and issuer URL from the IdP provider and populate the following environment variables:

 * DOMAIN_NAME
 * OPENID_CONNECT_CLIENT_ID
 * OPENID_CONNECT_CLIENT_SECRET
 * OPENID_CONNECT_ISSUER

(See above for details for each of these)

Note: Most IdP providers will require the terrareg installation to be accessed via https.
The instance should be configured with SSL certificates (SSL_CERT_PRIVATE_KEY/SSL_CERT_PUBLIC_KEY) or be hosted behind a reverse-proxy/load balancer.

### SAML2

Generate a public and a private key, using:

    openssl genrsa -out private.key 4096
    openssl req -new -x509 -key private.key -out publickey.cer -days 365

Set the folllowing environment variables:

* SAML2_IDP_METADATA_URL (required)
* SAML2_ENTITY_ID (required)
* SAML2_PRIVATE_KEY (required) (See above)
* SAML2_PUBLIC_KEY (required) (See above)
* SAML2_IDP_ENTITY_ID (optional)

In the IdP:

* configure the Single signin URL to `https://{terrareg_installation_domain}/saml/login?sso`;
* configure the request and response to be signed;
* ensure at least one attribute is assigned.