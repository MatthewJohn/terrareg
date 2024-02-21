# Upgrade

See the following notes for upgrading between major releases

## 3.0.0

To upgrade from v2.x.x to v3.0.0, the following need to be reviewed to perform the upgrade.

### Github authentication

The GITHUB_* environment variables have been replaced by a new [PROVIDER_SOURCES](docs/CONFIG.md#provider_sources) configuration.

To retain the same functionality, create a PROVIDER_SOURCES configuration with the following, replacing the variables with the values from your previous configurations:

```
PROVIDER_SOURCES='[{"name": "Github", "type": "github", "base_url": "https://github.com", "api_url": "https://api.github.com", "client_id": "$GITHUB_APP_CLIENT_ID", "client_secret": "$GITHUB_APP_CLIENT_SECRET", "app_id": "123456", "private_key_path": "$GITHUB_APP_PRIVATE_KEY_PATH", "webhook_secret": "$GITHUB_APP_WEBHOOK_SECRET", "auto_generate_namespaces": false}]'
```

After changing the environment variable, after the upgrade, Terrareg will generate a provider source for Github, which will allow authentication via Github.
