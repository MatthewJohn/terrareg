
import unittest.mock

import pytest

from terrareg.analytics import AnalyticsEngine
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
import terrareg.models
from test import client
from . import mock_record_module_version_download


class TestApiModuleVersionDownload(TerraregUnitTest):
    """Test ApiModuleVersionDownload resource."""

    @pytest.mark.parametrize('version', ['1.0.0', None])
    @setup_test_data()
    def test_existing_module_version_without_alaytics_token(self, version, client, mock_models):
        res = client.get(f"/v1/modules/testnamespace/testmodulename/testprovider/{f'{version}/' if version else ''}download")
        assert res.status_code == 401
        assert res.data == b'\nAn analytics token must be provided.\nPlease update module source to include analytics token.\n\nFor example:\n  source = "localhost/my-tf-application__testnamespace/testmodulename/testprovider"'

    @pytest.mark.parametrize('version', [
        # Test with version
        '0.1.2',
        # Test without version
        None
    ])
    @setup_test_data()
    def test_non_existent_module_version(self, version, client, mock_models):
        """Test endpoint with non-existent module"""
        res = client.get(f"/v1/modules/namespacename/modulename/testprovider/{f'{version}/' if version else ''}download")

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_existing_module_internal_download(self, client, mock_models, mock_record_module_version_download):
        """Test endpoint with analytics token"""

        res = client.get(
            '/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1/download',
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent'}
        )

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')
        test_module_version = terrareg.models.ModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='TestTerraformVersion',
            user_agent='TestUserAgent',
            auth_token=None
        )
        assert AnalyticsEngine.record_module_version_download.isinstance(
            AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'],
            terrareg.models.ModuleVersion
        )
        assert AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'].id == test_module_version.id

    @pytest.mark.parametrize('version,expected_version', [
        # Explicit latest version
        ('2.4.1', '2.4.1'),
        # Non-latest version
        ('1.0.0', '1.0.0'),
        # Latest endpoint
        (None, '2.4.1')
    ])
    @setup_test_data()
    def test_existing_module_internal_download_with_auth_token(
        self, version, expected_version, client, mock_models,
        mock_record_module_version_download):
        """Test endpoint with analytics token and auth token"""

        res = client.get(
            f"/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/{f'{version}/' if version else ''}download",
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent',
                     'Authorization': 'Bearer test123-authorization-token'}
        )

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')
        test_module_version = terrareg.models.ModuleVersion(module_provider=test_module_provider, version=expected_version)

        assert res.headers['X-Terraform-Get'] == f'/v1/terrareg/modules/testnamespace/testmodulename/testprovider/{expected_version}/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='TestTerraformVersion',
            user_agent='TestUserAgent',
            auth_token='test123-authorization-token'
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

        res = client.get(
            '/v1/modules/testnamespace/testmodulename/testprovider/2.4.1/download',
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent',
                     'Authorization': 'Bearer test123-authorization-token'}
        )
        assert res.status_code == 401

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @setup_test_data()
    def test_existing_module_download_with_internal_auth_token(
        self, client, mock_models,
        mock_record_module_version_download):
        """Test endpoint without analytics token and with an internal auth token"""

        with unittest.mock.patch('terrareg.config.Config.INTERNAL_EXTRACTION_ANALYITCS_TOKEN', 'unittest-internal-api-key'):
            res = client.get(
                '/v1/modules/testnamespace/testmodulename/testprovider/2.4.1/download',
                headers={'X-Terraform-Version': 'TestTerraformVersion',
                        'User-Agent': 'TestUserAgent',
                        'Authorization': 'Bearer unittest-internal-api-key'}
            )

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @setup_test_data()
    @pytest.mark.parametrize("module_url", [
        # Test without analytics token
        "/v1/modules/testnamespace/testmodulename/testprovider/2.4.1/download",

        # Test with analytics token
        "/v1/modules/test_analytics_token__testnamespace/testmodulename/testprovider/2.4.1/download"
    ])
    def test_existing_module_with_ignore_analytics_auth_token(self, module_url, client, mock_models, mock_record_module_version_download):
        """Test endpoint without analytics token and with an auth token to ignore analytics_token"""

        with unittest.mock.patch('terrareg.config.Config.IGNORE_ANALYTICS_TOKEN_AUTH_KEYS',
                                 ['unittest-ignore-analytics-auth-key', 'second-key']):
            res = client.get(
                module_url,
                headers={'X-Terraform-Version': 'TestTerraformVersion',
                        'User-Agent': 'TestUserAgent',
                        'Authorization': 'Bearer unittest-ignore-analytics-auth-key'}
            )

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_not_called()

    @setup_test_data()
    def test_existing_module_internal_download_with_invalid_auth_token_header(
        self, client, mock_models,
        mock_record_module_version_download):
        """Test endpoint with analytics token and auth token"""

        res = client.get(
            '/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1/download',
            headers={'X-Terraform-Version': 'TestTerraformVersion',
                     'User-Agent': 'TestUserAgent',
                     'Authorization': 'This is invalid'}
        )

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')
        test_module_version = terrareg.models.ModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.headers['X-Terraform-Get'] == '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            module_version=unittest.mock.ANY,
            analytics_token='test_token-name',
            terraform_version='TestTerraformVersion',
            user_agent='TestUserAgent',
            auth_token=None
        )
        assert AnalyticsEngine.record_module_version_download.isinstance(
            AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'],
            terrareg.models.ModuleVersion
        )
        assert AnalyticsEngine.record_module_version_download.call_args.kwargs['module_version'].id == test_module_version.id

