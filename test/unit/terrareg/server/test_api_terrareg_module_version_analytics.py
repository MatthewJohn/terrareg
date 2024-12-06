
import unittest.mock

import pytest

from terrareg.analytics import AnalyticsEngine
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
import terrareg.models
from test import client, mock_create_audit_event
from . import mock_record_module_version_download


class TestApiModuleVersionDownload(TerraregUnitTest):
    """Test ApiModuleVersionDownload resource."""

    @pytest.mark.parametrize("auth_token_prefix", [
        # No auth token
        "",
        None,

        # With example auth token
        "unittest-example-token",
    ])
    @setup_test_data()
    def test_existing_module_version_with_invalid_analytics_token(self, auth_token_prefix, client, mock_models, mock_record_module_version_download):
        """Test module version download with invalid analytics token"""
        with unittest.mock.patch("terrareg.config.Config.EXAMPLE_ANALYTICS_TOKEN", "unittest-example-token"):
            res = client.post(
                f"/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/1.0.0",
                json={
                    "analytics_token": auth_token_prefix,
                }
            )
        assert res.status_code == 400
        assert res.json == {
            "errors": ["Invalid analytics token"]
        }

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @pytest.mark.parametrize('auth_token_prefix', [
        # No auth token
        '',

        # With example auth token
        'unittest-example-token',

        # Valid analytics token
        'test'
    ])
    @setup_test_data()
    def test_existing_module_version_with_analytics_disabled(
            self, auth_token_prefix, client, mock_models, mock_record_module_version_download):
        """Test module version download endpoint with invalid analytics token and analytics disabled"""
        with unittest.mock.patch('terrareg.config.Config.EXAMPLE_ANALYTICS_TOKEN', 'unittest-example-token'), \
                unittest.mock.patch('terrareg.config.Config.DISABLE_ANALYTICS', True):

            res = client.post(
                f"/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/2.4.1",
            )

        assert res.status_code == 400
        assert res.json == {
            "errors": ["Analytics is disabled"]
        }

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @pytest.mark.parametrize('version', [
        # Test with version
        '0.1.2',
        # Test without version
        None
    ])
    @setup_test_data()
    def test_non_existent_module_version(self, version, client, mock_models, mock_record_module_version_download):
        """Test endpoint with non-existent module"""
        res = client.post(f"/v1/terrareg/analytics/namespacename/modulename/testprovider/{version}")

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404
        AnalyticsEngine.record_module_version_download.assert_not_called()

    @setup_test_data()
    def test_existing_module(self, client, mock_models, mock_record_module_version_download):
        """Test endpoint with analytics token"""

        res = client.post(
            '/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/2.4.1',
            headers={
                'X-Terraform-Version': 'TestTerraformVersion',
                'User-Agent': 'TestUserAgent'
            },
            json={
                "analytics_token": "test_token-name",
                "terraform_version": "some-terraform-version"
            }
        )

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')
        test_module_version = terrareg.models.ModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.status_code == 200

        AnalyticsEngine.record_module_version_download.assert_called_with(
            namespace_name='testnamespace',
            module_name='testmodulename',
            provider_name='testprovider',
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='some-terraform-version',
            user_agent=None,
            auth_token=None,
            ignore_user_agent=True
        )
        assert AnalyticsEngine.record_module_version_download.isinstance(
            AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'],
            terrareg.models.ModuleVersion
        )
        assert AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'].id == test_module_version.id

    @setup_test_data()
    def test_existing_module_internal_download_with_auth_token_without_analytics_token(
        self, client, mock_models,
        mock_record_module_version_download):
        """Test endpoint with valid auth token and without analytics token"""

        with unittest.mock.patch('terrareg.config.Config.INTERNAL_EXTRACTION_ANALYTICS_TOKEN', 'unittest-internal-api-key'):
            res = client.post(
                '/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/2.4.1',
                headers={
                    'X-Terraform-Version': 'TestTerraformVersion',
                    'User-Agent': 'TestUserAgent',
                    'Authorization': 'Bearer unittest-internal-api-key'
                },
                json={
                }
            )
            assert res.status_code == 200

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @setup_test_data()
    def test_existing_module_download_with_internal_auth_token(
        self, client, mock_models,
        mock_record_module_version_download):
        """Test endpoint without analytics token and with an internal auth token"""

        with unittest.mock.patch('terrareg.config.Config.INTERNAL_EXTRACTION_ANALYTICS_TOKEN', 'unittest-internal-api-key'):
            res = client.post(
                '/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/2.4.1',
                headers={
                    'Authorization': 'Bearer unittest-internal-api-key'
                },
                json={
                    "analytics_token": "testtoken"
                }
            )

        assert res.status_code == 200

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @setup_test_data()
    def test_existing_module_with_ignore_analytics_auth_token(self, client, mock_models, mock_record_module_version_download):
        """Test endpoint without analytics token and with an auth token to ignore analytics_token"""

        with unittest.mock.patch('terrareg.config.Config.IGNORE_ANALYTICS_TOKEN_AUTH_KEYS',
                                ['unittest-ignore-analytics-auth-key', 'second-key']):
            res = client.post(
                "/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/2.4.1",
                headers={
                    'Authorization': 'Bearer unittest-ignore-analytics-auth-key',
                },
                json={
                    "analytics_token": "test_analytics_token",
                    "terraform_version": "TestTerraformVersion",
                }
            )

        assert res.status_code == 200

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @pytest.mark.parametrize('auth_token', [
        'ignore-analytics-token',
        'analytics-auth-token',
        'internal-extraction-token'
    ])
    @setup_test_data()
    def test_required_authentication(self, auth_token, client, mock_models, mock_record_module_version_download):
        """Test that various forms of authentication work when unauthenticated access is disabled"""

        with unittest.mock.patch('terrareg.config.Config.ANALYTICS_AUTH_KEYS', ['analytics-auth-token:dev']), \
                unittest.mock.patch('terrareg.config.Config.IGNORE_ANALYTICS_TOKEN_AUTH_KEYS', ['ignore-analytics-token']), \
                unittest.mock.patch('terrareg.config.Config.INTERNAL_EXTRACTION_ANALYTICS_TOKEN', 'internal-extraction-token'), \
                unittest.mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', False), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_PRESIGNED_URL_SECRET', 'test'):
            res = client.post(
                f"/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/2.4.1",
                headers={
                    'Authorization': f'Bearer {auth_token}'
                },
                json={
                    "analytics_token": "test"
                }
            )
        assert res.status_code == 200

    @setup_test_data()
    def test_unauthenticated(self, client, mock_models, mock_record_module_version_download):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.post('/v1/terrareg/analytics/testnamespace/testmodulename/testprovider/1.0.0')

        self._test_unauthenticated_terraform_api_endpoint_test(call_endpoint)
