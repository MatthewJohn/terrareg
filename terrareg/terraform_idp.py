from datetime import datetime, timedelta
import json
import os
import time
import uuid

from jwkest.jwk import RSAKey, rsa_load
from pyop.authz_state import AuthorizationState
from pyop.provider import Provider
from pyop.subject_identifier import HashBasedSubjectIdentifierFactory
from pyop.userinfo import Userinfo
import sqlalchemy

import terrareg.config
from terrareg.constants import TERRAFORM_REDIRECT_URI_PORT_RANGE
import terrareg.auth
from terrareg.database import Database


class TerraformIdpUserLookup:
    """Implement pyop.userinfo.Userinfo to provide interface for looking up users"""

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


class BaseIdpDatabase:

    SHOULD_EXPIRE = False

    @property
    def table(self):
        """Return table for database"""
        raise NotImplementedError

    def __getitem__(self, item):
        """Obtain data from database where code matches the key"""
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    self.table.c.data
                ).select_from(
                    self.table
                ).where(
                    self.table.c.key==item
                )
            )
            item = res.first()
            if item:
                return json.loads(Database.decode_blob(item[0]))
            raise KeyError

    def __setitem__(self, key, value):
        """Set code and data into database"""
        db = Database.get()
        with db.get_connection() as conn:
            # Perform update, determine if any rows are affected and fallback
            # to insert.
            # Upsert operations appear to be database-specific and might
            # be brittle across SQLite and MySQL
            blob_value = Database.encode_blob(json.dumps(value))
            with conn.begin() as transaction:
                res = conn.execute(
                    sqlalchemy.update(
                        self.table
                    ).where(
                        self.table.c.key==key
                    ).values(
                        data=blob_value
                    )
                )
                if res.rowcount == 0:
                    res = conn.execute(
                        sqlalchemy.insert(
                            self.table
                        ).values(
                            key=key,
                            data=blob_value,
                            expiry=(datetime.now() + timedelta(seconds=terrareg.config.Config().TERRAFORM_OIDC_IDP_SESSION_EXPIRY))
                        )
                    )


                if self.SHOULD_EXPIRE:
                    # Delete any old sessions
                    conn.execute(sqlalchemy.delete(self.table).where(self.table.c.expiry < datetime.now()))
                    transaction.commit()


    def __contains__(self, item):
        """Determine if code exists"""
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    sqlalchemy.func.count(self.table.c.id)
                ).select_from(
                    self.table
                ).where(
                    self.table.c.key==item
                )
            )
            return bool(res.scalar())

    def items(self):
        """Return all code/data from database"""
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    self.table.c.key,
                    self.table.c.data
                ).select_from(
                    self.table
                )
            )
            return [
                [row[0], json.loads(Database.decode_blob(row[1]))]
                for row in res.all()
            ]


class AuthorizationCodeDatabase(BaseIdpDatabase):

    SHOULD_EXPIRE = True

    @property
    def table(self):
        """Return table"""
        return Database.get().terraform_idp_authorization_code


class AccessTokenDatabase(BaseIdpDatabase):

    SHOULD_EXPIRE = True

    @property
    def table(self):
        """Return table"""
        return Database.get().terraform_idp_access_token


class SubjectIdentifierDatabase(BaseIdpDatabase):

    @property
    def table(self):
        """Return table"""
        return Database.get().terraform_idp_subject_identifier


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

        if config.ALLOW_UNAUTHENTICATED_ACCESS:
            if config.DEBUG:
                print("Disabling Terraform OIDC provider due to ALLOW_UNAUTHENTICATED_ACCESS is true")
            return False

        if not config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT:
            if config.DEBUG:
                print('Disabling Terraform OIDC provider due to missing TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT')
            return False

        if not os.path.isfile(config.TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH):
            if config.DEBUG:
                print('Disabling Terraform OIDC provider due to TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH not set to a present file')
            return False

        if not config.PUBLIC_URL:
            if config.DEBUG:
                print('Disabling Terraform OIDC provider due to missing PUBLIC_URL config')
            return False
        return True

    def __init__(self):
        """Check config and store member variables"""
        self._provider = None

    @property
    def provider(self):
        """Obtain singleton instance of provider"""
        if self._provider is None:
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
            self._provider = Provider(
                signing_key=signing_key,
                configuration_information=configuration_information,
                authz_state=AuthorizationState(
                    HashBasedSubjectIdentifierFactory(terrareg.config.Config().TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT),
                    authorization_code_db=AuthorizationCodeDatabase(),
                    access_token_db=AccessTokenDatabase(),
                    subject_identifier_db=SubjectIdentifierDatabase(),
                    authorization_code_lifetime=terrareg.config.Config().TERRAFORM_OIDC_IDP_SESSION_EXPIRY,
                    access_token_lifetime=terrareg.config.Config().TERRAFORM_OIDC_IDP_SESSION_EXPIRY,
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
                userinfo=userinfo_db,
                id_token_lifetime=terrareg.config.Config().TERRAFORM_OIDC_IDP_SESSION_EXPIRY
            )
        return self._provider
