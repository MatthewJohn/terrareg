
import unittest.mock

import pytest

from test import BaseTest

from .base_terraform_static_token_tests import BaseTerraformStaticTokenTests
from terrareg.auth import TerraformAnalyticsAuthKeyAuthMethod


class TestTerraformAnalyticsAuthKeyAuthMethod(BaseTerraformStaticTokenTests):

    CLS = TerraformAnalyticsAuthKeyAuthMethod

    @pytest.mark.parametrize('analaytics_auth_keys, expected_result', [
        ([], False),
        ([':dev', ':prod'], False),
        (['', ''], False),
        (['abcefg:dev'], True),
        (['abcefg:dev', ''], True),
        (['abcefg:dev', 'efghy:prod'], True),
        (['', 'abcefg:dev'], True),
    ])
    def test_is_enabled(self, analaytics_auth_keys, expected_result):
        """Test get_username method"""
        obj = self.CLS()
        with unittest.mock.patch('terrareg.config.Config.ANALYTICS_AUTH_KEYS', analaytics_auth_keys):
            assert obj.is_enabled() is expected_result

    @pytest.mark.parametrize('analaytics_auth_keys, authorization_header, expected_result', [
        ([], None, False),
        ([':dev', ':prod'], None, False),
        (['', ''], None, False),
        (['abcefg:dev'], None, False),
        (['abcefg:dev', ''], None, False),
        (['abcefg:dev', 'efghy:prod'], None, False),
        (['', 'abcefg:dev'], None, False),

        # Invalid headers
        (['abcefg:dev'], 'abcefg', False),
        (['abcefg:dev'], '', False),
        # Incorrect match
        (['abcefg:dev'], 'Bearer nomatch', False),
        # Partial matches
        (['abcefg:dev'], 'Bearer abcef', False),
        (['abcefg:dev'], 'Bearer bcefg', False),

        # Match
        (['abcefg:dev'], 'Bearer abcefg', True),
        (['zxcvb:dev', 'abcefg:prod'], 'Bearer abcefg', True),
        (['zxcvb:dev', 'abcefg:prod'], 'Bearer zxcvb', True),
    ])
    def test_check_auth_state(self, analaytics_auth_keys, authorization_header, expected_result):
        """Test get_username method"""
        obj = self.CLS()
        headers = {}
        if authorization_header is not None:
            headers['Authorization'] = authorization_header

        with unittest.mock.patch('terrareg.config.Config.ANALYTICS_AUTH_KEYS', analaytics_auth_keys), \
                BaseTest.get().SERVER._app.test_request_context(headers=headers) as request_context:
            assert obj.check_auth_state() is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = self.CLS()
        assert obj.get_username() == "Terraform deployment analytics token"

    def test_should_record_terraform_analytics(self):
        """Test should_record_terraform_analytics method"""
        obj = self.CLS()
        assert obj.should_record_terraform_analytics() is True

    @pytest.mark.parametrize('authorization_header, expected_result', [
        (None, None),
        ('', None),
        ('baretoken', None),
        ('Bearerjoined', None),
        ('Bearer abcef', "abcef"),
    ])
    def test_get_terraform_auth_token(self, authorization_header, expected_result):
        """Test get_terraform_auth_token method"""
        obj = self.CLS()
        headers = {}
        if authorization_header is not None:
            headers['Authorization'] = authorization_header

        with BaseTest.get().SERVER._app.test_request_context(headers=headers):
            assert obj.get_terraform_auth_token() == expected_result
