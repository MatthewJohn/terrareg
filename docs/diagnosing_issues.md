
# Diagnosing issues

## The domain shown in Terraform snippets contains the wrong URL

This can happen if the following are true:

 * [PUBLIC_URL](./CONFIG.md#public_url) has not been set
 * [DOMAIN_NAME](./CONFIG.md#domain_name) (deprecated) has not been set
 * The domain is rewritten by a reverse proxy before traffic reaches Terrareg

To fix this, set [PUBLIC_URL](./CONFIG.md#public_url) to the URL that users access the Terrareg instance.
