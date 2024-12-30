
from datetime import datetime, timedelta
import os
from tempfile import NamedTemporaryFile
from unittest import mock
import pytest
import sqlalchemy
import pyop.authz_state
import pyop.subject_identifier

from terrareg.terraform_idp import AccessTokenDatabase, AuthorizationCodeDatabase, SubjectIdentifierDatabase, TerraformIdp, TerraformIdpUserLookup
from terrareg.database import Database
from test.integration.terrareg import TerraregIntegrationTest


class TestAuthorizationCodeDatabase(TerraregIntegrationTest):

    def setup_method(self, method):
        """Delete all data before method"""
        super().setup_method(method)
        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(sqlalchemy.delete(database.terraform_idp_authorization_code))

    def test_insert(self):
        """Test inserting data"""
        db = AuthorizationCodeDatabase()

        # Ensure value does not already exist
        with pytest.raises(KeyError):
            db['testcode']
        
        db['testcode'] = {'this': 'that'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_authorization_code.c.key,
                    database.terraform_idp_authorization_code.c.data
                ).select_from(database.terraform_idp_authorization_code)
            ).all()
            assert row_data == [
                ("testcode", b'{"this": "that"}')
            ]

        
    def test_upserting_data(self):
        """Test updating value in database"""
        db = AuthorizationCodeDatabase()

        # Ensure value does not already exist
        with pytest.raises(KeyError):
            db['testcode']
        
        db['testcode'] = {'this': 'that'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_authorization_code.c.key,
                    database.terraform_idp_authorization_code.c.data
                ).select_from(database.terraform_idp_authorization_code)
            ).all()
            assert row_data == [
                ("testcode", b'{"this": "that"}')
            ]

        # Update data
        db['testcode'] = {'this2': 'that2'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_authorization_code.c.key,
                    database.terraform_idp_authorization_code.c.data
                ).select_from(database.terraform_idp_authorization_code)
            ).all()
            assert row_data == [
                ("testcode", b'{"this2": "that2"}')
            ]

    def test_read_data(self):
        """Test data expiry"""
        db = AuthorizationCodeDatabase()

        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(
                sqlalchemy.insert(
                    database.terraform_idp_authorization_code
                ).values(
                    key="non-expired-key",
                    data=b'{"real": "value"}',
                    expiry=(datetime.now() + timedelta(minutes=5))
                )
            )

        # Insert second key, which should tidy up any expired keys
        db['second-key'] = {}

        # Attempt to get data
        assert db['non-expired-key'] == {"real": "value"}

    def test_data_expiry(self):
        """Test data expiry"""
        db = AuthorizationCodeDatabase()

        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(
                sqlalchemy.insert(
                    database.terraform_idp_authorization_code
                ).values(
                    key="expired-key",
                    data=b'{"old": "value"}',
                    expiry=(datetime.now() - timedelta(minutes=5))
                )
            )

        # Insert second key, which should tidy up any expired keys
        db['second-key'] = {}

        # Attempt to get data
        with pytest.raises(KeyError):
            db['expired-key']

    def test_list_all_data(self):
        """Test getting all data"""
        db = AuthorizationCodeDatabase()

        db['test1'] = {'test': 1}
        db['test2'] = {'test': 2}


        items = db.items()
        assert len(items) == 2
        # Test individually as order may not be consistent between database types
        assert ['test1', {'test': 1}] in items
        assert ['test2', {'test': 2}] in items

    def test_contains(self):
        """Test contains"""

        db = AuthorizationCodeDatabase()

        db['test-exists'] = {'test': 2}

        assert 'test-exists' in db
        assert 'test-doesnot-exist' not in db


class TestAccessTokenDatabase(TerraregIntegrationTest):

    def setup_method(self, method):
        """Delete all data before method"""
        super().setup_method(method)
        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(sqlalchemy.delete(database.terraform_idp_access_token))

    def test_insert(self):
        """Test inserting data"""
        db = AccessTokenDatabase()

        # Ensure value does not already exist
        with pytest.raises(KeyError):
            db['testcode']
        
        db['testcode'] = {'this': 'that'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_access_token.c.key,
                    database.terraform_idp_access_token.c.data
                ).select_from(database.terraform_idp_access_token)
            ).all()
            assert row_data == [
                ("testcode", b'{"this": "that"}')
            ]

        
    def test_upserting_data(self):
        """Test updating value in database"""
        db = AccessTokenDatabase()

        # Ensure value does not already exist
        with pytest.raises(KeyError):
            db['testcode']
        
        db['testcode'] = {'this': 'that'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_access_token.c.key,
                    database.terraform_idp_access_token.c.data
                ).select_from(database.terraform_idp_access_token)
            ).all()
            assert row_data == [
                ("testcode", b'{"this": "that"}')
            ]

        # Update data
        db['testcode'] = {'this2': 'that2'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_access_token.c.key,
                    database.terraform_idp_access_token.c.data
                ).select_from(database.terraform_idp_access_token)
            ).all()
            assert row_data == [
                ("testcode", b'{"this2": "that2"}')
            ]

    def test_read_data(self):
        """Test data expiry"""
        db = AccessTokenDatabase()

        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(
                sqlalchemy.insert(
                    database.terraform_idp_access_token
                ).values(
                    key="non-expired-key",
                    data=b'{"real": "value"}',
                    expiry=(datetime.now() + timedelta(minutes=5))
                )
            )

        # Insert second key, which should tidy up any expired keys
        db['second-key'] = {}

        # Attempt to get data
        assert db['non-expired-key'] == {"real": "value"}

    def test_data_expiry(self):
        """Test data expiry"""
        db = AccessTokenDatabase()

        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(
                sqlalchemy.insert(
                    database.terraform_idp_access_token
                ).values(
                    key="expired-key",
                    data=b'{"old": "value"}',
                    expiry=(datetime.now() - timedelta(minutes=5))
                )
            )

        # Insert second key, which should tidy up any expired keys
        db['second-key'] = {}

        # Attempt to get data
        with pytest.raises(KeyError):
            db['expired-key']

    def test_list_all_data(self):
        """Test getting all data"""
        db = AccessTokenDatabase()

        db['test1'] = {'test': 1}
        db['test2'] = {'test': 2}


        items = db.items()
        assert len(items) == 2
        # Test individually as order may not be consistent between database types
        assert ['test1', {'test': 1}] in items
        assert ['test2', {'test': 2}] in items

    def test_contains(self):
        """Test contains"""

        db = AccessTokenDatabase()

        db['test-exists'] = {'test': 2}

        assert 'test-exists' in db
        assert 'test-doesnot-exist' not in db



class TestSubjectIdentifierDatabase(TerraregIntegrationTest):

    def setup_method(self, method):
        """Delete all data before method"""
        super().setup_method(method)
        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(sqlalchemy.delete(database.terraform_idp_subject_identifier))

    def test_insert(self):
        """Test inserting data"""
        db = SubjectIdentifierDatabase()

        # Ensure value does not already exist
        with pytest.raises(KeyError):
            db['testcode']
        
        db['testcode'] = {'this': 'that'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_subject_identifier.c.key,
                    database.terraform_idp_subject_identifier.c.data
                ).select_from(database.terraform_idp_subject_identifier)
            ).all()
            assert row_data == [
                ("testcode", b'{"this": "that"}')
            ]

    def test_upserting_data(self):
        """Test updating value in database"""
        db = SubjectIdentifierDatabase()

        # Ensure value does not already exist
        with pytest.raises(KeyError):
            db['testcode']
        
        db['testcode'] = {'this': 'that'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_subject_identifier.c.key,
                    database.terraform_idp_subject_identifier.c.data
                ).select_from(database.terraform_idp_subject_identifier)
            ).all()
            assert row_data == [
                ("testcode", b'{"this": "that"}')
            ]

        # Update data
        db['testcode'] = {'this2': 'that2'}

        database = Database.get()
        with database.get_connection() as conn:
            row_data = conn.execute(
                sqlalchemy.select(
                    database.terraform_idp_subject_identifier.c.key,
                    database.terraform_idp_subject_identifier.c.data
                ).select_from(database.terraform_idp_subject_identifier)
            ).all()
            assert row_data == [
                ("testcode", b'{"this2": "that2"}')
            ]

    def test_read_data(self):
        """Test data expiry"""
        db = SubjectIdentifierDatabase()

        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(
                sqlalchemy.insert(
                    database.terraform_idp_subject_identifier
                ).values(
                    key="non-expired-key",
                    data=b'{"real": "value"}',
                    expiry=(datetime.now() + timedelta(minutes=5))
                )
            )

        # Insert second key, which should tidy up any expired keys
        db['second-key'] = {}

        # Attempt to get data
        assert db['non-expired-key'] == {"real": "value"}

    def test_data_non_expiry(self):
        """Test data does not expiry"""
        db = SubjectIdentifierDatabase()

        database = Database.get()
        with database.get_connection() as conn:
            conn.execute(
                sqlalchemy.insert(
                    database.terraform_idp_subject_identifier
                ).values(
                    key="expired-key",
                    data=b'{"old": "value"}',
                    expiry=(datetime.now() - timedelta(minutes=5))
                )
            )

        # Insert second key, which should tidy up any expired keys
        db['second-key'] = {}

        # Attempt to get data
        assert db['expired-key'] == {"old": "value"}

    def test_list_all_data(self):
        """Test getting all data"""
        db = SubjectIdentifierDatabase()

        db['test1'] = {'test': 1}
        db['test2'] = {'test': 2}


        items = db.items()
        assert len(items) == 2
        # Test individually as order may not be consistent between database types
        assert ['test1', {'test': 1}] in items
        assert ['test2', {'test': 2}] in items

    def test_contains(self):
        """Test contains"""

        db = SubjectIdentifierDatabase()

        db['test-exists'] = {'test': 2}

        assert 'test-exists' in db
        assert 'test-doesnot-exist' not in db


class TestTerraformIdp(TerraregIntegrationTest):
    """Test TerraformIdp class"""

    @pytest.mark.parametrize('allow_unauthenticated_access_config, terraform_oidc_idp_subject_id_hash_salt_config, '
                             'terraform_oidc_idp_signing_key_path_config, terraform_oidc_idp_signing_key_path_exists, '
                             'public_url_config, expected_result', [
        # Working example
        (False, 'somesecret', '/tmp/keyfile-unittest', True, 'https://localhost', True),

        # Public access enabled
        (True, 'somesecret', '/tmp/keyfile-unittest', True, 'https://localhost', False),

        # Without secret
        (False, None, '/tmp/keyfile-unittest', True, 'https://localhost', False),

        # File doesn't exist
        (False, 'somsecret', '/tmp/keyfile-unittest', False, 'https://localhost', False),

        # Public URL not configured
        (False, 'somsecret', '/tmp/keyfile-unittest', True, None, False),
    ])
    def test_is_enabled(self, allow_unauthenticated_access_config,
                        terraform_oidc_idp_subject_id_hash_salt_config,
                        terraform_oidc_idp_signing_key_path_config,
                        terraform_oidc_idp_signing_key_path_exists,
                        public_url_config,
                        expected_result):

        # Create empty file, if sign key should exist
        if terraform_oidc_idp_signing_key_path_exists:
            with open(terraform_oidc_idp_signing_key_path_config, "w") as fh:
                fh.write("")

        try:
            with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access_config), \
                    mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT',
                            terraform_oidc_idp_subject_id_hash_salt_config), \
                    mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH',
                            terraform_oidc_idp_signing_key_path_config), \
                    mock.patch('terrareg.config.Config.PUBLIC_URL', public_url_config):

                assert TerraformIdp.get().is_enabled == expected_result

        finally:
            # Always clear up file
            if terraform_oidc_idp_signing_key_path_exists:
                os.remove(terraform_oidc_idp_signing_key_path_config)

    def test_provider_properties(self):
        """Check provider property"""
        # Signing RSA key
        signing_rsa_key = """
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDg9lttk9fpB7+PxpjVfZZPUC0NT8VGzzaT2qJlbyafY7HNPyBr
ixGc/EZbwx73FYhFnGW0IQd8xxTqlBZOFoAbI9Kx850J1J+gGn3IUbW3dm9aQq0d
cwMuhrMj45Ixiwd14cyGb+ZFsmGpdqRAEM2nbeQEnA5eNre0/uVGNuR+CQIDAQAB
AoGAdmk2NrdbLo2lh0hBqh4wwA6zqA4VCPCJCcpLMJkQ+1S+ggp4RiMtYjRn1GUg
J25uDDYGUooQJt2jZNYN54xwYNwXobGaCSlmWSfGfiCF6SKlVICf+d8EEYa8GcAM
rBDyTMghayn0oA03loSdAG5iqzF1ob/zQXgNCPJkc2C/IAECQQDwWRK2gt12edPh
kYr8XD9Hakjs8EaNEB4xO8GKCmnLhjRZDvMj5usXGkSfPo24qutssyYpn/nP6YR0
1/Q0mcNRAkEA75zI91DU82fMHhct2GgfEP2IvdaHHQ8zZnarC9Prn+6/6cNefhtN
S0+tiZj0R0B3dkLGTTqcmYSQe/EEjY2xOQJBAJnR9+b0s/W6HH91nUTLaPg0rn1t
fUmUci5CNyg4Z+MIfgItTjDA/d4oQpjD+QGh6dAEi70CFGga5Fm/SBxN+DECQBBV
7A2QYTRG+0+B3QpH7vZFkrD+ky+T/bkalga0Z/f7WvIg86w9SEO+JuKenujMqFhT
rRlOyaZdt0v73oeYBWECQQDc7n98Cx6G1Nt2/87o6UaYzW5N4SfWCPTaiS9/inpQ
yzEmVAlL/QfgkKm+0zsa8czkSwNjtBz9vOIffCxtZmlf
-----END RSA PRIVATE KEY-----
""".strip()
        signing_key_path = NamedTemporaryFile().name
        with open(signing_key_path, "w") as signing_key_fh:
            signing_key_fh.write(signing_rsa_key)

        try:
            with mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH', signing_key_path), \
                    mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT', "unittest-hash-salt"), \
                    mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SESSION_EXPIRY', 1523), \
                    mock.patch('terrareg.config.Config.PUBLIC_URL', "https://example.local"):
                provider = TerraformIdp.get().provider


            # Check signing key
            ## Ensure private key was loaded correctly
            assert provider.signing_key.thumbprint(hash_function="SHA-512").hex() == "c96cef84767032ddcfa7de269ad67b4859fc3eb7f7c93bfd7d51cf1716350ee1be98f98ff0402e8fe6a3742e9c22825b9e222fe9600bcd5b1975044031a87dab"
            ## Ensure signing key algorithm is correct
            assert provider.signing_key.alg == "RS256"

            # Check configuration information
            assert provider.configuration_information._dict == {
                'authorization_endpoint': 'https://example.local/terraform/oauth/authorization',
                'backchannel_logout_session_supported': False,
                'backchannel_logout_supported': False,
                'claims_parameter_supported': True,
                'frontchannel_logout_session_supported': False,
                'frontchannel_logout_supported': False,
                'grant_types_supported': ['authorization_code', 'implicit'],
                'id_token_signing_alg_values_supported': ['RS256'],
                'issuer': 'https://example.local',
                'jwks_uri': 'https://example.local/terraform/oauth/jwks',
                'request_parameter_supported': False,
                'request_uri_parameter_supported': True,
                'require_request_uri_registration': False,
                'response_modes_supported': ['query',
                'fragment'],
                'response_types_supported': ['code', 'code id_token', 'code token', 'code id_token token'],
                'scopes_supported': ['openid'],
                'subject_types_supported': ['pairwise'],
                'token_endpoint': 'https://example.local/terraform/oauth/token',
                'token_endpoint_auth_methods_supported': ['client_secret_basic'],
                'version': '3.0',
            }

            # Check authz state
            assert isinstance(provider.authz_state, pyop.authz_state.AuthorizationState)

            # Check authz state hash subject identifier factory
            assert isinstance(provider.authz_state._subject_identifier_factory, pyop.subject_identifier.HashBasedSubjectIdentifierFactory)
            assert provider.authz_state._subject_identifier_factory.hash_salt == "unittest-hash-salt"

            # Check authz state authorization code DB
            assert isinstance(provider.authz_state.authorization_codes, AuthorizationCodeDatabase)

            # Check authz state access token DB
            assert isinstance(provider.authz_state.access_tokens, AccessTokenDatabase)

            # Check authz state subject identifier DB
            assert isinstance(provider.authz_state.subject_identifiers, SubjectIdentifierDatabase)

            # Check authz state authorization code lifetime
            assert provider.authz_state.authorization_code_lifetime == 1523

            # Check clients
            assert provider.clients == {
                'terraform-cli': {
                    'client_id': 'terraform-cli',
                    # "Issued at" is irelevant, copy from
                    # actual data
                    'client_id_issued_at': provider.clients.get("terraform-cli", {}).get("client_id_issued_at", 0),
                    'redirect_uris': [
                        'http://localhost:10000/login',
                        'http://localhost:10001/login',
                        'http://localhost:10002/login',
                        'http://localhost:10003/login',
                        'http://localhost:10004/login',
                        'http://localhost:10005/login',
                        'http://localhost:10006/login',
                        'http://localhost:10007/login',
                        'http://localhost:10008/login',
                        'http://localhost:10009/login',
                        'http://localhost:10010/login'
                    ],
                    'response_types': ['code'],
                    'token_endpoint_auth_method': 'none'
                }
            }

            # Check user info DB
            assert isinstance(provider.userinfo, TerraformIdpUserLookup)

            # Check ID token lifetime
            assert provider.id_token_lifetime == 1523

        finally:
            os.unlink(signing_key_path)

