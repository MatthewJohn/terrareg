
from datetime import datetime, timedelta
import os
from unittest import mock
import pytest
import sqlalchemy

from terrareg.terraform_idp import AccessTokenDatabase, AuthorizationCodeDatabase, SubjectIdentifierDatabase, TerraformIdp
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
