import random
import string

import requests
from oauthlib.oauth2 import WebApplicationClient

import terrareg.config

class OpenidConnect:

    _METADATA_CONFIG = None

    @classmethod
    def is_enabled(cls):
        """Whether OpenID connect authentication is enabled"""
        config = terrareg.config.Config()
        return bool(config.OPENID_CONNECT_CLIENT_ID and config.OPENID_CONNECT_CLIENT_SECRET and config.OPENID_CONNECT_ISSUER and config.DOMAIN_NAME)

    def get_client():
        """Return oauth2 web application client"""
        return WebApplicationClient(terrareg.config.Config().OPENID_CONNECT_CLIENT_ID)

    @staticmethod
    def get_redirect_url():
        """Obtain redirect URL for Terrareg instance"""
        config = terrareg.config.Config()
        return f'https://{config.DOMAIN_NAME}/openid/callback'

    @classmethod
    def obtain_issuer_metadata(cls):
        """Obtain wellknown metadata from issuer"""
        # Obtain meta data from well-known URL, if not previously cached
        if not cls.is_enabled():
            return None

        if cls._METADATA_CONFIG is None:
            res = requests.get(terrareg.config.Config().OPENID_CONNECT_ISSUER + '/.well-known/openid-configuration')
            cls._METADATA_CONFIG = res.json()

        return cls._METADATA_CONFIG

    @classmethod
    def generate_state(cls):
        """Return random string for state"""
        letters = string.ascii_letters + string.digits
        return ''.join(random.choice(letters) for i in range(24))

    @classmethod
    def get_authorize_redirect_url(cls):
        """Get authorize URL to redirect user to for authentication"""
        if not cls.is_enabled:
            return None, None

        auth_url = cls.obtain_issuer_metadata().get('authorization_endpoint', None)
        if not auth_url:
            return None, None

        config = terrareg.config.Config()

        state = cls.generate_state()

        return cls.get_client().prepare_request_uri(
            auth_url,
            redirect_uri=cls.get_redirect_url(),
            scope=['openid', 'profile'],
            state=state
        ), state

    @classmethod
    def fetch_access_token(cls, code):
        """Fetch access token from OpenID issuer"""
        client = cls.get_client()
        config = terrareg.config.Config()
        token_request_body = client.prepare_request_body(
            code=code,
            client_id=config.OPENID_CONNECT_CLIENT_ID,
            client_secret=config.OPENID_CONNECT_CLIENT_SECRET,
            redirect_uri=cls.get_redirect_url()
        )

        token_endpoint = cls.obtain_issuer_metadata().get('token_endpoint', None)
        if not token_endpoint:
            return None
        response = requests.post(token_endpoint, token_request_body)
        print(response.text)

        return client.parse_request_body_response(response.text)


