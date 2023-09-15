import os
import time
import uuid

from jwkest.jwk import RSAKey, rsa_load
from pyop.authz_state import AuthorizationState
from pyop.provider import Provider
from pyop.subject_identifier import HashBasedSubjectIdentifierFactory
from pyop.userinfo import Userinfo

import terrareg.config
from terrareg.constants import TERRAFORM_REDIRECT_URI_PORT_RANGE
import terrareg.auth


class TerraformIdpUserLookup:
    """Implement pypo.userinfo.Userinfo to provide interface for looking up users"""

    def __init__(self):
        pass

    def __getitem__(self, item):
        auth_method = terrareg.auth.AuthFactory().get_current_auth_method()
        return {
            "sub": auth_method.get_username()
        }

    def __contains__(self, item):
        """Always lookup valid user"""
        return True

    def get_claims_for(self, user_id, requested_claims, userinfo=None):
        """
        Terraform does not request any claims, so immediately return
        """
        return {}


class MockDB:
    """Implement pypo.userinfo.Userinfo to provide interface for looking up users"""

    def __init__(self, name):
        self.name = name
        self.db = {}

    def __getitem__(self, item):
        print(f"{self.name}: {item}")
        return self.db.get(item)

    def __setitem__(self, key, value):
        print(f"{self.name}: Setting {key} to {value}")
        self.db[key] = value

    def __contains__(self, item):
        """Always lookup valid user"""
        print(f"{self.name}: {item}")
        return item in self.db

    def items(self):
        return self.db.items()

    def get_claims_for(self, user_id, requested_claims, userinfo=None):
        """
        Terraform does not request any claims, so immediately return
        """
        return {}


class TerraformIdp:
    """Handle creating of IDP objects using pyop"""

    _INSTANCE = None

    @classmethod
    def get(cls):
        """Get singleton instance of class"""
        if cls._INSTANCE is None:
            cls._INSTANCE = cls()
        return cls._INSTANCE

    @property
    def is_enabled(self):
        """Whether the provider is enabled"""
        config = terrareg.config.Config()

        if not config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT:
            print('Disabling Terraform OIDC provider due to missing TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT')
            return False

        if not os.path.isfile(config.TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH):
            print('Disabling Terraform OIDC provider due to TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH not set to a present file')
            return False

        if not config.PUBLIC_URL:
            print('Disabling Terraform OIDC provider due to missing PUBLIC_URL config')
            return False
        return True

    def __init__(self):
        """Check config and store member variables"""
        config = terrareg.config.Config()

        issuer = config.PUBLIC_URL
        oauth_base_url = f"{config.PUBLIC_URL}/terraform/oauth"
        authentication_endpoint = f"{oauth_base_url}/authorization"
        jwks_uri = f"{oauth_base_url}/jwks"
        token_endpoint = f"{oauth_base_url}/token"

        configuration_information = {
            'issuer': issuer,
            'authorization_endpoint': authentication_endpoint,
            'jwks_uri': jwks_uri,
            'token_endpoint': token_endpoint,
            'scopes_supported': ['openid'],
            'response_types_supported': ['code', 'code id_token', 'code token', 'code id_token token'],
            'response_modes_supported': ['query', 'fragment'],
            'grant_types_supported': ['authorization_code', 'implicit'],
            'subject_types_supported': ['pairwise'],
            'token_endpoint_auth_methods_supported': ['client_secret_basic'],
            'claims_parameter_supported': True
        }

        userinfo_db = TerraformIdpUserLookup()
        signing_key = RSAKey(key=rsa_load(terrareg.config.Config().TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH), alg='RS256')
        self.provider = Provider(
            signing_key=signing_key,
            configuration_information=configuration_information,
            authz_state=AuthorizationState(
                HashBasedSubjectIdentifierFactory(terrareg.config.Config().TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT),
                authorization_code_db=MockDB("authorization_code_db"),
                access_token_db=MockDB("access_token_db"),
                subject_identifier_db=MockDB("subject_identifier_db")
            ),
            clients={
                "terraform-cli": {
                    "response_types": ["code"],
                    # Match client ID from Terraform well-known endpoint
                    "client_id": "terraform-cli",
                    "client_id_issued_at": int(time.time()),
                    "token_endpoint_auth_method": 'none',
                    # Match redirect URLs provided by terraform well-known endpoint
                    "redirect_uris": [
                        f"http://localhost:{port}/login"
                        for port in range(TERRAFORM_REDIRECT_URI_PORT_RANGE[0], TERRAFORM_REDIRECT_URI_PORT_RANGE[1] + 1)
                    ]
                }
            },
            userinfo=userinfo_db
        )
