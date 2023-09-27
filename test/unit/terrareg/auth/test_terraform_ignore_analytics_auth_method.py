
import unittest.mock

import pytest

from test import BaseTest

from .base_terraform_static_token_tests import BaseTerraformStaticTokenTests
from terrareg.auth import TerraformIgnoreAnalyticsAuthMethod


class TestTerraformIgnoreAnalyticsAuthMethod(BaseTerraformStaticTokenTests):

    CLS = TerraformIgnoreAnalyticsAuthMethod

    @pytest.mark.parametrize('ignore_analytics_auth_tokens, expected_result', [
        ([], False),
        ([''], False),
        (['', ''], False),
        (['abcefg'], True),
        (['abcefg', ''], True),
        (['abcefg', 'efghy'], True),
        (['', 'abcefg'], True),
    ])
    def test_is_enabled(self, ignore_analytics_auth_tokens, expected_result):
        """Test get_username method"""
        obj = self.CLS()
        with unittest.mock.patch('terrareg.config.Config.IGNORE_ANALYTICS_TOKEN_AUTH_KEYS', ignore_analytics_auth_tokens):
            assert obj.is_enabled() is expected_result

    @pytest.mark.parametrize('ignore_analytics_auth_tokens, authorization_header, expected_result', [
        ([], None, False),
        (['', ''], None, False),
        (['abcefg'], None, False),
        (['abcefg', ''], None, False),
        (['abcefg', 'efghy'], None, False),
        (['', 'abcefg'], None, False),

        # Invalid headers
        (['abcefg'], 'abcefg', False),
        (['abcefg'], '', False),
        # Incorrect match
        (['abcefg'], 'Bearer nomatch', False),
        # Partial matches
        (['abcefg'], 'Bearer abcef', False),
        (['abcefg'], 'Bearer bcefg', False),

        # Match
        (['abcefg'], 'Bearer abcefg', True),
        (['zxcvb', 'abcefg'], 'Bearer abcefg', True),
        (['zxcvb', 'abcefg'], 'Bearer zxcvb', True),
    ])
    def test_check_auth_state(self, ignore_analytics_auth_tokens, authorization_header, expected_result):
        """Test get_username method"""
        obj = self.CLS()
        headers = {}
        if authorization_header is not None:
            headers['Authorization'] = authorization_header

        with unittest.mock.patch('terrareg.config.Config.IGNORE_ANALYTICS_TOKEN_AUTH_KEYS', ignore_analytics_auth_tokens), \
                BaseTest.get().SERVER._app.test_request_context(headers=headers) as request_context:
            assert obj.check_auth_state() is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = self.CLS()
        assert obj.get_username() == "Terraform ignore analytics token"

    def test_should_record_terraform_analytics(self):
        """Test should_record_terraform_analytics method"""
        obj = self.CLS()
        assert obj.should_record_terraform_analytics() is False

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
