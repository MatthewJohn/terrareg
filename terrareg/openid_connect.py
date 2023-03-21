import datetime
import random
import string
import json

import jwt
import requests
import oauthlib.oauth2

import terrareg.config
from terrareg.utils import get_public_url_details


class OpenidConnect:

    _METADATA_CONFIG = None

    _JWKS_CLIENT = None

    @classmethod
    def is_enabled(cls):
        """Whether OpenID connect authentication is enabled"""
        config = terrareg.config.Config()
        _, domain, _ = get_public_url_details()
        return bool(config.OPENID_CONNECT_CLIENT_ID and config.OPENID_CONNECT_CLIENT_SECRET and config.OPENID_CONNECT_ISSUER and domain)

    def get_client():
        """Return oauth2 web application client"""
        return oauthlib.oauth2.WebApplicationClient(terrareg.config.Config().OPENID_CONNECT_CLIENT_ID)

    @staticmethod
    def get_redirect_url():
        """Obtain redirect URL for Terrareg instance"""
        _, domain, _ = get_public_url_details()
        return f'https://{domain}/openid/callback'

    @classmethod
    def get_jwks_client(cls):
        """Obtain instance of jwks_client"""
        if not cls._JWKS_CLIENT:
            jwks_uri = cls.obtain_issuer_metadata().get('jwks_uri', None)
            if jwks_uri is None:
                raise Exception("No jwks_uri found")

            cls._JWKS_CLIENT = jwt.PyJWKClient(jwks_uri, cache_keys=True)

        return cls._JWKS_CLIENT

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

        state = cls.generate_state()

        return cls.get_client().prepare_request_uri(
            auth_url,
            redirect_uri=cls.get_redirect_url(),
            scope=['openid', 'profile'],
            state=state
        ), state

    @classmethod
    def fetch_access_token(cls, uri, valid_state):
        """Fetch access token from OpenID issuer"""
        client = cls.get_client()
        config = terrareg.config.Config()

        callback_response = client.parse_request_uri_response(uri=uri, state=valid_state)

        token_request_body = client.prepare_request_body(
            code=callback_response.get('code'),
            client_id=config.OPENID_CONNECT_CLIENT_ID,
            client_secret=config.OPENID_CONNECT_CLIENT_SECRET,
            redirect_uri=cls.get_redirect_url()
        )

        metadata = cls.obtain_issuer_metadata()
        if metadata is None:
            return None

        token_endpoint = metadata.get('token_endpoint', None)
        if not token_endpoint:
            return None

        response = requests.post(
            token_endpoint,
            token_request_body,
            headers={'Content-Type': 'application/x-www-form-urlencoded'})

        return client.parse_request_body_response(response.text)

    @classmethod
    def validate_session_token(cls, session_id_token):
        """Validate session token, ensuring it is valid"""
        header = jwt.get_unverified_header(jwt=session_id_token)
        key = cls.get_jwks_client().get_signing_key(header["kid"])

        jwt.decode(
            session_id_token,
            key=key.key,
            algorithms=[header['alg']],
            audience=terrareg.config.Config().OPENID_CONNECT_CLIENT_ID
        )

    @classmethod
    def get_user_info(cls, access_token):
        """Get user infor"""
        user_info_endpoint = cls.obtain_issuer_metadata().get('userinfo_endpoint')
        if not user_info_endpoint:
            return None

        res = requests.post(
            user_info_endpoint,
            headers={
                'Authorization': f'Bearer {access_token}'
            }
        )
        return res.json()
