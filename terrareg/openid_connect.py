
import requests
from oauthlib.oauth2 import WebApplicationClient

import terrareg.config

class OpenidConnection:

    def is_enabled(self):
        """Whether OpenID connect authentication is enabled"""
        config = terrareg.config.Config()
        return bool(config.OPENID_CONNECT_CLIENT_ID and config.OPENID_CONNECT_CLIENT_SECRET and config.OPENID_CONNECT_AUTH_URL and config.DOMAIN_NAME)

    def get_client():
        """Return oauth2 web application client"""
        return WebApplicationClient(terrareg.config.Config().OPENID_CONNECT_CLIENT_ID)

    @classmethod
    def get_authorize_url(cls):
        """Get authorize URL to redirect user to for authentication"""
        config = terrareg.config.Config()
        return cls.get_client().prepare_request_uri(
            config.OPENID_CONNECT_AUTH_URL,
            redirect_uri = f'https://{config.DOMAIN_NAME}/openid/callback',
            scope = ['read:user'],
            state = 'D8VAo311AAl_49LAtM51HA'
        )

    @classmethod
    def fetch_access_token():
        pass
