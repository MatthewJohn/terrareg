
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


class TestModuleProviderPageTerraformDownload(TerraregUnitTest):
    """Test /modules/ web endpoint, testing ApiModuleVersionDownload resource when terraform-get is provided."""

    @pytest.mark.parametrize('version', ['1.0.0', 'latest', None])
    @setup_test_data()
    def test_existing_module_version_without_alaytics_token(self, version, client, mock_models):
        res = client.get(f"/modules/testnamespace/testmodulename/testprovider{f'/{version}' if version else ''}?terraform-get=1")
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
        res = client.get(f"modules/namespacename/modulename/testprovider{f'/{version}' if version else ''}?terraform-get=1")

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_existing_module_internal_download(self, client, mock_models, mock_record_module_version_download):
        """Test endpoint with analytics token"""

        res = client.get(
            '/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1?terraform-get=1',
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
            namespace_name='testnamespace',
            module_name='testmodulename',
            provider_name='testprovider',
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

    @pytest.mark.parametrize('namespace,module,provider,version,expected_version,expected_return_url', [
        ## Archive download
        # Explicit latest version
        ('testnamespace', 'testmodulename', 'testprovider', '2.4.1', '2.4.1', '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'),
        # Non-latest version
        ('testnamespace', 'testmodulename', 'testprovider', '1.0.0', '1.0.0', '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/1.0.0/source.zip'),
        # Latest endpoint
        ('testnamespace', 'testmodulename', 'testprovider', 'latest', '2.4.1', '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'),
        ('testnamespace', 'testmodulename', 'testprovider', None, '2.4.1', '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/source.zip'),

        ## Git provider
        ('moduleextraction', 'gitextraction', 'usesgitproviderwithversions', '2.2.2', '2.2.2', 'git::ssh://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions?ref=v2.2.2'),
        ('moduleextraction', 'gitextraction', 'usesgitproviderwithversions', '2.1.0', '2.1.0', 'git::ssh://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions?ref=v2.1.0'),
        ('moduleextraction', 'gitextraction', 'usesgitproviderwithversions', 'latest', '2.2.2', 'git::ssh://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions?ref=v2.2.2'),
        ('moduleextraction', 'gitextraction', 'usesgitproviderwithversions', None, '2.2.2', 'git::ssh://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions?ref=v2.2.2'),

        ## Custom git URL
        ('moduleextraction', 'gitextraction', 'placeholdercloneurl', '5.2.3', '5.2.3', 'git::ssh://git@localhost:7999/moduleextraction/gitextraction-placeholdercloneurl.git?ref=v5.2.3'),
        ('moduleextraction', 'gitextraction', 'placeholdercloneurl', '4.0.0', '4.0.0', 'git::ssh://git@localhost:7999/moduleextraction/gitextraction-placeholdercloneurl.git?ref=v4.0.0'),
        ('moduleextraction', 'gitextraction', 'placeholdercloneurl', 'latest', '5.2.3', 'git::ssh://git@localhost:7999/moduleextraction/gitextraction-placeholdercloneurl.git?ref=v5.2.3'),
        ('moduleextraction', 'gitextraction', 'placeholdercloneurl', None, '5.2.3', 'git::ssh://git@localhost:7999/moduleextraction/gitextraction-placeholdercloneurl.git?ref=v5.2.3'),
    ])
    @setup_test_data()
    def test_existing_module_internal_download_with_auth_token(
        self, namespace, module, provider, version, expected_version, expected_return_url, client, mock_models,
        mock_record_module_version_download):
        """Test endpoint with analytics token and auth token"""

        with unittest.mock.patch('terrareg.config.Config.ANALYTICS_AUTH_KEYS', ['test123-authorization-token:dev']):
            res = client.get(
                f"/modules/test_token-name__{namespace}/{module}/{provider}{f'/{version}' if version else ''}?terraform-get=1",
                headers={'X-Terraform-Version': 'TestTerraformVersion',
                        'User-Agent': 'TestUserAgent',
                        'Authorization': 'Bearer test123-authorization-token'}
            )

        test_namespace = terrareg.models.Namespace(name=namespace)
        test_module = terrareg.models.Module(namespace=test_namespace, name=module)
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name=provider)
        test_module_version = terrareg.models.ModuleVersion(module_provider=test_module_provider, version=expected_version)

        assert res.headers['X-Terraform-Get'] == expected_return_url
        assert res.status_code == 204

        AnalyticsEngine.record_module_version_download.assert_called_with(
            namespace_name=namespace,
            module_name=module,
            provider_name=provider,
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

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        # @TODO Test authentication from Terraform
        def call_endpoint():
            return client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
