
import pytest

from terrareg.models import Module, ModuleProvider, Namespace
from test.integration.terrareg import TerraregIntegrationTest

class TestProviderLogo(TerraregIntegrationTest):

    @pytest.mark.parametrize('provider_name,expect_exists', [
        ('aws', True),
        ('gcp', True),
        ('null', True),
        ('doesnotexist', False),
    ])
    def test_logo_exists(self, provider_name, expect_exists):
        """Test exists method of logo provider"""
        namespace = Namespace(name='real_providers')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider(module=module, name=provider_name)
        logo = module_provider.get_logo()
        assert logo.exists == expect_exists

    @pytest.mark.parametrize('provider_name,expected_tos', [
        ('aws', 'Amazon Web Services, AWS, the Powered by AWS logo are trademarks of Amazon.com, Inc. or its affiliates.'),
        ('gcp', 'Google Cloud and the Google Cloud logo are trademarks of Google LLC.'),
        ('null', ' '),
        ('doesnotexist', None),
    ])
    def test_logo_tos(self, provider_name, expected_tos):
        """Test tos property of ProviderLogo"""
        namespace = Namespace(name='real_providers')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider(module=module, name=provider_name)
        logo = module_provider.get_logo()
        assert logo.tos == expected_tos

    @pytest.mark.parametrize('provider_name,expected_alt', [
        ('aws', 'Powered by AWS Cloud Computing'),
        ('gcp', 'Google Cloud'),
        ('null', 'Null Provider'),
        ('doesnotexist', None),
    ])
    def test_logo_alt(self, provider_name, expected_alt):
        """Test alt property of ProviderLogo"""
        namespace = Namespace(name='real_providers')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider(module=module, name=provider_name)
        logo = module_provider.get_logo()
        assert logo.alt == expected_alt

    @pytest.mark.parametrize('provider_name,expected_link', [
        ('aws', 'https://aws.amazon.com/'),
        ('gcp', 'https://cloud.google.com/'),
        ('null', '#'),
        ('doesnotexist', None),
    ])
    def test_logo_link(self, provider_name, expected_link):
        """Test link property of ProviderLogo"""
        namespace = Namespace(name='real_providers')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider(module=module, name=provider_name)
        logo = module_provider.get_logo()
        assert logo.link == expected_link

    @pytest.mark.parametrize('provider_name,expected_source', [
        ('aws', '/static/images/PB_AWS_logo_RGB_stacked.547f032d90171cdea4dd90c258f47373c5573db5.png'),
        ('gcp', '/static/images/gcp.png'),
        ('null', '/static/images/null.png'),
        ('doesnotexist', None),
    ])
    def test_logo_source(self, provider_name, expected_source):
        """Test source property of ProviderLogo"""
        namespace = Namespace(name='real_providers')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider(module=module, name=provider_name)
        logo = module_provider.get_logo()
        assert logo.source == expected_source
