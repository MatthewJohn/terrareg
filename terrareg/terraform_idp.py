import os
import time
import uuid

from jwkest.jwk import RSAKey, rsa_load
from pyop.authz_state import AuthorizationState
from pyop.provider import Provider
from pyop.subject_identifier import HashBasedSubjectIdentifierFactory
from pyop.userinfo import Userinfo

import terrareg.config


class TerraformIdpUserLookup:
    """Implement pypo.userinfo.Userinfo to provide interface for looking up users"""
    def __init__(self, db):
        self._db = db

    def __getitem__(self, item):
        return self._db[item]

    def __contains__(self, item):
        return item in self._db

    def get_claims_for(self, user_id, requested_claims, userinfo=None):
        """
        Filter the userinfo based on which claims where requested.
        :param user_id: user identifier
        :param requested_claims: see <a href="http://openid.net/specs/openid-connect-core-1_0.html#ClaimsParameter">
            "OpenID Connect Core 1.0", Section 5.5</a> for structure
        :param userinfo: if user_info is specified the claims will be filtered from the user_info directly instead
        first querying the storage against the user_id
        :return: All requested claims available from the userinfo.
        """

        if not userinfo:
            userinfo = self._db[user_id] if user_id else {}
        claims = {claim: userinfo[claim] for claim in requested_claims if claim in userinfo}
        return claims


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
        userinfo_endpoint = f"{oauth_base_url}/userinfo"
        registration_endpoint = f"{oauth_base_url}/registration"
        end_session_endpoint = f"{oauth_base_url}/logout"
        print(issuer)

        configuration_information = {
            'issuer': issuer,
            'authorization_endpoint': authentication_endpoint,
            'jwks_uri': jwks_uri,
            'token_endpoint': token_endpoint,
            'userinfo_endpoint': userinfo_endpoint,
            # 'registration_endpoint': registration_endpoint,
            'end_session_endpoint': end_session_endpoint,
            'scopes_supported': ['openid'],
            'response_types_supported': ['code', 'code id_token', 'code token', 'code id_token token'],
            'response_modes_supported': ['query', 'fragment'],
            'grant_types_supported': ['authorization_code', 'implicit'],
            'subject_types_supported': ['pairwise'],
            'token_endpoint_auth_methods_supported': ['client_secret_basic'],
            'claims_parameter_supported': True
        }

        userinfo_db = Userinfo({})
        signing_key = RSAKey(key=rsa_load(terrareg.config.Config().TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH), alg='RS256')
        self.provider = Provider(
            signing_key=signing_key,
            configuration_information=configuration_information,
            authz_state=AuthorizationState(HashBasedSubjectIdentifierFactory(terrareg.config.Config().TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT)),
            clients={
                "terraform-cli": {
                    "response_types": ["code"],
                    # Match client ID from Terraform well-known endpoint
                    "client_id": "terraform-cli",
                    "client_id_issued_at": int(time.time()),
                    "client_secret": uuid.uuid4().hex,
                    "client_secret_expires_at": 0,  # never expires
                }
            },
            userinfo=userinfo_db
        )
