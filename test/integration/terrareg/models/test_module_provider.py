
from unittest import mock
import pytest

from terrareg.models import Module, Namespace, ModuleProvider
import terrareg.errors
from test.integration.terrareg import (
    TerraregIntegrationTest,
    setup_test_data
)


class TestModuleProvider(TerraregIntegrationTest):

    @pytest.mark.parametrize('module_provider_name', [
        'invalid@atsymbol',
        'invalid"doublequote',
        "invalid'singlequote",
        '-startwithdash',
        'endwithdash-',
        '_startwithunderscore',
        'endwithunscore_',
        'a:colon',
        'or;semicolon',
        'who?knows',
        'with-dash',
        'with_underscore',
        'withAcapital',
        'StartwithCaptital',
        'endwithcapitaL',
        ''
    ])
    def test_invalid_module_provider_names(self, module_provider_name):
        """Test invalid module names"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        with pytest.raises(terrareg.errors.InvalidModuleProviderNameError):
            ModuleProvider(module=module, name=module_provider_name)

    @pytest.mark.parametrize('module_provider_name', [
        'normalname',
        'name2withnumber',
        '2startendiwthnumber2',
        'contains4number'
    ])
    def test_valid_module_provider_names(self, module_provider_name):
        """Test valid module names"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        ModuleProvider(module=module, name=module_provider_name)


    def test_module_provider_name_in_allow_list(self):
        """Test module provider name that is not in allow list"""
        with mock.patch('terrareg.models.ALLOWED_PROVIDERS', ['aws', 'azure', 'test']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            ModuleProvider(module=module, name='aws')
            ModuleProvider(module=module, name='azure')
            ModuleProvider(module=module, name='test')


    def test_module_provider_name_not_in_allow_list(self):
        """Test module provider name that is not in allow list"""
        with mock.patch('terrareg.models.ALLOWED_PROVIDERS', ['onlyallow']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            with pytest.raises(terrareg.errors.ProviderNameNotPermittedError):
                ModuleProvider(module=module, name='notallowed')

    @setup_test_data()
    def test_module_provider_get_versions(self):
        """Test that a module provider with versions in the wrong order are still returned correctly."""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        assert [mv.version for mv in module_provider.get_versions()] == [
            '10.23.0', '2.1.0', '1.5.4',
            '0.1.10', '0.1.09', '0.1.8',
            '0.1.1', '0.0.9'
        ]
