
from unittest import mock
import pytest

from terrareg.models import Module, ModuleVersion, Namespace, ModuleProvider
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
        with mock.patch('terrareg.config.ALLOWED_PROVIDERS', ['aws', 'azure', 'test']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            ModuleProvider(module=module, name='aws')
            ModuleProvider(module=module, name='azure')
            ModuleProvider(module=module, name='test')


    def test_module_provider_name_not_in_allow_list(self):
        """Test module provider name that is not in allow list"""
        with mock.patch('terrareg.config.ALLOWED_PROVIDERS', ['onlyallow']):
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

    @setup_test_data()
    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/suffix'),
        ('no-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/suffix'),
        ('module-provider-urls', '1.2.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/suffix'),
        ('module-provider-urls', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

    @setup_test_data()
    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, None),
        ('no-git-provider', '1.4.0', 'unittestpath', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/suffix'),
        ('module-provider-urls', '1.2.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/suffix'),
        ('module-provider-urls', '1.4.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url_with_custom_module_version_urls_disabled(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values when module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        with mock.patch('terrareg.config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

    @setup_test_data()
    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/suffix'),
        ('no-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', None, 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/suffix'),
        ('git-provider-urls', '1.1.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.1.0/unittestpathsuffix'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/suffix'),
        ('git-provider-urls', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/git-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', None, None),
        ('module-provider-urls', '1.2.0', 'unittestpath', None),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/suffix'),
        ('module-provider-urls', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.4.0/unittestpathsuffix'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url_with_custom_module_provider_urls_disabled(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values when module provider repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        with mock.patch('terrareg.config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url
