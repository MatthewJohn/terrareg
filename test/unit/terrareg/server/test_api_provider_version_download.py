
from datetime import datetime
import unittest.mock

import pytest

import terrareg.analytics
from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.models
import terrareg.provider_search
import terrareg.provider_model
import terrareg.provider_version_model


class TestApiProviderVersionDownload(TerraregIntegrationTest):
    """Test ApiProviderVersionDownload endpoint"""

    def test_endpoint(self, client):
        """Test endpoint."""
        mock_record_provider_version_download = unittest.mock.MagicMock(side_effect=terrareg.analytics.ProviderAnalytics.record_provider_version_download)
        with unittest.mock.patch('terrareg.analytics.ProviderAnalytics.record_provider_version_download', mock_record_provider_version_download):
            res = client.get('/v1/providers/initial-providers/multiple-versions/1.5.0/download/linux/amd64')

        assert res.status_code == 200
        assert res.json == {
            'arch': 'amd64',
            'download_url': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions/releases/download/v1.5.0/terraform-provider-multiple-versions_1.5.0_linux_amd64.zip',
            'filename': 'terraform-provider-multiple-versions_1.5.0_linux_amd64.zip',
            'os': 'linux',
            'protocols': ['5.0'],
            'shasum': 'a26d0401981bf2749c129ab23b3037e82bd200582ff7489e0da2a967b50daa98',
            'shasums_signature_url': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions/releases/download/v1.5.0/terraform-provider-multiple-versions_1.5.0_SHA256SUMS.sig',
            'shasums_url': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions/releases/download/v1.5.0/terraform-provider-multiple-versions_1.5.0_SHA256SUMS',
            'signing_keys': {
                'gpg_public_keys': [
                    {
                        'ascii_armor': '-----BEGIN PGP PUBLIC '
                                        'KEY BLOCK-----\n'
                                        '\n'
                                        'mI0EZUHt7QEEAKgSXXCkqShvE54omLsE0Gzu/Es2Nelwnps8ETlcHPKag0VlZch/\n'
                                        '0HPyF3hGsdZM7GB1il7fGCGw6Urkmci7XkRj2M09QtAvE2YPOqfNfMvHQrLIAkBV\n'
                                        'lP/4xIBnGMmsUYVMAeo0DiDdFf3Q3pIbWDhd7+OCPKh80F/pYM1Rm4qnABEBAAG0\n'
                                        'UVRlc3QgVGVycmFyZWcgVGVzdHMgKFRlc3QgS2V5IGZvciB0ZXJyYXJlZyBUZXN0\n'
                                        'cykgPHRlcnJhcmVnLXRlc3RzQGNvbGFtYWlsLmNvLnVrPojOBBMBCgA4FiEEIadO\n'
                                        'Tj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwMFCwkIBwIGFQoJCAsCBBYCAwECHgEC\n'
                                        'F4AACgkQNN43SsNkDNtkywP/SR8U/c3gzAY4w0KF3ZG5sBJqrBfdA2d2R//Bsjvz\n'
                                        'jRCpGdaXVBJG2FFyfl5QLLhC56rS6nsX6vcXkrRGQtYG6Bhroo6eWjVnyT1RMM+A\n'
                                        'wD5uwCijPlSdl82q91aFQk3jwqNoe4/gr9ERHagx3MAgMTEhIzPaKpGHtL7TPM+B\n'
                                        'nOi4jQRlQe3tAQQAxCeKNhBAv13aXeSvPI1JKW9pcg5g9Hfd4s/qj82/0hE/Kfgt\n'
                                        '4u7RGOEe7q1WgKirtoiv/XSpwKMSlXtt9AH8lbgkveiJ3V+DqJxdzCm42Zlyvg9Z\n'
                                        '9sqLz6XOAyMkv44U1x182KMipuuethRmSemN8jthc4Bh5iEM/l7460IyRk8AEQEA\n'
                                        'AYi2BBgBCgAgFiEEIadOTj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwwACgkQNN43\n'
                                        'SsNkDNtn+AP+Pm3+u+if0BExYTMKJ0/dU4ICWBkyuuMDkQlz8oOn9/w9EYvkqR/r\n'
                                        'QypRou1K0KbLxBCz0vqAM7KLXe0rKwZZ3eWSThiwTJkFlkJsUgwMqqROteYmWm3S\n'
                                        'MK0hMLszB/mfN0Q2DW4U0tWslehdEA+aaccwA5PVFKdkA12ImK500TY=\n'
                                        '=EL4W\n'
                                        '-----END PGP PUBLIC KEY '
                                        'BLOCK-----',
                        'key_id': '34DE374AC3640CDB',
                        'source': '',
                        'source_url': None,
                        'trust_signature': ''
                    }
                ]
            }
        }

        namespace = terrareg.models.Namespace.get(name='initial-providers')
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name='multiple-versions')
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version='1.5.0')

        mock_record_provider_version_download.assert_called_once_with(
            namespace_name='initial-providers',
            provider_name='multiple-versions',
            provider_version=provider_version,
            terraform_version=None,
            user_agent='werkzeug/2.2.3'
        )

    def test_endpoint_with_provider_invalid_architecture(self, client):
        """Test endpoint with invalid architecture"""
        res = client.get('/v1/providers/initial-providers/multiple-versions/1.5.0/download/linux/powerpc')
        assert res.status_code == 404

    def test_endpoint_with_provider_invalid_os(self, client):
        """Test endpoint with invalid operating system"""
        res = client.get('/v1/providers/initial-providers/multiple-versions/1.5.0/download/dos/amd64')
        assert res.status_code == 404

    def test_endpoint_with_provider_non_existent_binary(self, client):
        """Test endpoint with invalid arch/os that doesn't exist in the provider version"""
        res = client.get('/v1/providers/initial-providers/multiple-versions/1.5.0/download/windows/arm64')
        assert res.status_code == 404

    def test_endpoint_non_existent_provider_version(self, client):
        """Test endpoint with non-existent provider version"""
        res = client.get('/v1/providers/initial-providers/multiple-versions/1.9.20/download/windows/amd64')
        assert res.status_code == 404

    def test_endpoint_non_existent_provider(self, client):
        """Test endpoint with non-existent provider"""
        res = client.get('/v1/providers/initial-providers/doesnotexist/1.9.20/download/windows/amd64')
        assert res.status_code == 404

    def test_endpoint_non_existent_namespace(self, client):
        """Test endpoint with non-existent namespace"""
        res = client.get('/v1/providers/doesnotexist/doesnotexist/1.9.20/download/windows/amd64')
        assert res.status_code == 404

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/providers/initial-providers/multiple-versions/1.5.0/download/linux/amd64')

        self._test_unauthenticated_terraform_api_endpoint_test(call_endpoint)
