
import unittest.mock

import pytest

from test import BaseTest

from .base_terraform_static_token_tests import BaseTerraformStaticTokenTests
from terrareg.auth import TerraformInternalExtractionAuthMethod


class TestTerraformInternalExtractionAuthMethod(BaseTerraformStaticTokenTests):

    CLS = TerraformInternalExtractionAuthMethod

    @pytest.mark.parametrize('internal_extraction_token, expected_result', [
        (None, False),
        ('', False),
        ('isavalue', True),
    ])
    def test_is_enabled(self, internal_extraction_token, expected_result):
        """Test get_username method"""
        obj = self.CLS()
        with unittest.mock.patch('terrareg.config.Config.INTERNAL_EXTRACTION_ANALYITCS_TOKEN', internal_extraction_token):
            assert obj.is_enabled() is expected_result

    @pytest.mark.parametrize('internal_extraction_token, authorization_header, expected_result', [
        (None, None, False),
        ('', None, False),
        ('abcefg', None, False),

        # Invalid headers
        ('abcefg', 'abcefg', False),
        ('abcefg', '', False),
        # Incorrect match
        ('abcefg', 'Bearer nomatch', False),
        # Partial matches
        ('abcefg', 'Bearer abcef', False),
        ('abcefg', 'Bearer bcefg', False),

        # Match
        ('abcefg', 'Bearer abcefg', True),
    ])
    def test_check_auth_state(self, internal_extraction_token, authorization_header, expected_result):
        """Test get_username method"""
        obj = self.CLS()
        headers = {}
        if authorization_header is not None:
            headers['Authorization'] = authorization_header

        with unittest.mock.patch('terrareg.config.Config.INTERNAL_EXTRACTION_ANALYITCS_TOKEN', internal_extraction_token), \
                BaseTest.get().SERVER._app.test_request_context(headers=headers) as request_context:
            assert obj.check_auth_state() is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = self.CLS()
        assert obj.get_username() == "Terraform internal extraction"

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
