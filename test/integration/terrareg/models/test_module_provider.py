
from unittest import mock
import pytest

from terrareg.models import Module, ModuleVersion, Namespace, ModuleProvider
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest


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
        with mock.patch('terrareg.config.Config.ALLOWED_PROVIDERS', ['aws', 'azure', 'test']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            ModuleProvider(module=module, name='aws')
            ModuleProvider(module=module, name='azure')
            ModuleProvider(module=module, name='test')


    def test_module_provider_name_not_in_allow_list(self):
        """Test module provider name that is not in allow list"""
        with mock.patch('terrareg.config.Config.ALLOWED_PROVIDERS', ['onlyallow']):
            namespace = Namespace(name='test')
            module = Module(namespace=namespace, name='test')
            with pytest.raises(terrareg.errors.ProviderNameNotPermittedError):
                ModuleProvider(module=module, name='notallowed')

    def test_module_provider_get_versions(self):
        """Test that a module provider with versions in the wrong order are still returned correctly."""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        assert [mv.version for mv in module_provider.get_versions()] == [
            '23.2.3-beta', '10.23.0', '2.1.0',
            '1.5.4', '0.1.10', '0.1.09', '0.1.8',
            '0.1.1', '0.0.9'
        ]

    def test_module_provider_get_versions_without_beta(self):
        """Test that a module provider with versions in the wrong order are still returned correctly."""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        assert [mv.version for mv in module_provider.get_versions(include_beta=False)] == [
            '10.23.0', '2.1.0', '1.5.4',
            '0.1.10', '0.1.09', '0.1.8',
            '0.1.1', '0.0.9'
        ]

    def test_module_provider_get_latest_version(self):
        """
        Test that a module provider with versions in the wrong order return correct
        latest version and ignores beta version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.get_latest_version()

        assert module_version.version == '10.23.0'

    @pytest.mark.parametrize('module_name', [
        # Module with no versions at all
        'noversions',
        # Module with only unpublished version
        'onlyunpublished',
        # Module with only a published beta version
        'onlybeta'
    ])
    def test_module_provider_get_latest_version_with_no_version(self, module_name):
        """
        Test that a module provider without any versions does not return
        a latest version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.get_latest_version()

        assert module_version is None

    def test_module_provider_calculate_latest_version(self):
        """
        Test that a module provider with versions in the wrong order return correct
        latest version and ignores beta version with calculate_latest_version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='wrongversionorder')
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.calculate_latest_version()

        assert module_version.version == '10.23.0'

    @pytest.mark.parametrize('module_name', [
        # Module with no versions at all
        'noversions',
        # Module with only unpublished version
        'onlyunpublished',
        # Module with only a published beta version
        'onlybeta'
    ])
    def test_module_provider_calculate_latest_version_with_no_version(self, module_name):
        """
        Test that a module provider without any versions does not return
        a latest version using calculate_latest_version.
        """
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='testprovider')
        module_version = module_provider.calculate_latest_version()

        assert module_version is None

    @pytest.mark.parametrize('module_name,module_version,path,expected_browse_url', [
        # Test no browse URL in any configuration
        ('no-git-provider', '1.0.0', None, None),
        ('no-git-provider', '1.0.0', 'unittestpath', None),
        # Test browse URL only configured in module version
        ('no-git-provider', '1.4.0', None, 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/suffix'),
        ('no-git-provider', '1.4.0', 'unittestpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-test/browse/1.4.0/unittestpathsuffix'),

        # Test URI encoded tags in browse URL template
        # - from module provider
        ('no-git-provider-uri-encoded', '1.4.0', 'testpath', 'https://mv-browse-url.com/repo_url_tests/no-git-provider-uri-encoded-test/browse/testpath?at=release%40test%2F1.4.0%2F'),
        # - from git provider
        ('git-provider-uri-encoded', '1.4.0', 'testpath', 'https://browse-url.com/repo_url_tests/git-provider-uri-encoded-test/browse/testpath?at=release%40test%2F1.4.0%2F'),

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

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None, 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/suffix'),
        ('module-provider-urls', '1.2.0', 'unittestpath', 'https://mp-browse-url.com/repo_url_tests/module-provider-urls-test/browse/1.2.0/unittestpathsuffix'),
        # Test with repo URls configured in module provider and override in module version
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
        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

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

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None, None),
        ('module-provider-urls', '1.2.0', 'unittestpath', None),
        # Test with repo URls configured in module provider and override in module version
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
        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

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

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None, None),
        ('module-provider-urls', '1.2.0', 'unittestpath', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None, None),
        ('module-provider-urls', '1.4.0', 'unittestpath', None),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', None, 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/suffix'),
        ('module-provider-override-git-provider', '1.3.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.3.0/unittestpathsuffix'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', None, 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/suffix'),
        ('module-provider-override-git-provider', '1.4.0', 'unittestpath', 'https://browse-url.com/repo_url_tests/module-provider-override-git-provider-test/browse/1.4.0/unittestpathsuffix')
    ])
    def test_get_source_browse_url_with_custom_module_provider_and_module_version_urls_disabled(self, module_name, module_version, path, expected_browse_url):
        """Ensure browse URL matches the expected values when module provider and module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        kwargs = {'path': path} if path else {}
        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
                assert module_version.get_source_browse_url(**kwargs) == expected_browse_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://mp-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url_with_custom_module_version_urls_disabled(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values when module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://mv-clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url_with_custom_module_provider_urls_disabled(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values when module provider repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_clone_url', [
        # Test no clone URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test clone URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'ssh://clone-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'ssh://clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'ssh://clone-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_clone_url_with_custom_module_provider_and_module_version_urls_disabled(self, module_name, module_version, expected_clone_url):
        """Ensure clone URL matches the expected values when module provider and module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
                assert module_version.get_git_clone_url() == expected_clone_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with git provider configured in module provider
        ('module-provider-urls', '1.2.0', 'https://mp-base-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://mp-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', 'https://mp-base-url.com/repo_url_tests/module-provider-urls-test'),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'https://mp-base-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://mp-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://mp-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url_with_custom_module_version_urls_disabled(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values when module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
            assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/no-git-provider-test'),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-urls-test'),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://mv-base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url_with_custom_module_provider_urls_disabled(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values when module provider repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            assert module_version.get_source_base_url() == expected_base_url

    @pytest.mark.parametrize('module_name,module_version,expected_base_url', [
        # Test no base URL in any configuration
        ('no-git-provider', '1.0.0', None),
        # Test base URL only configured in module version
        ('no-git-provider', '1.4.0', None),

        # Test with git provider configured in module provider
        ('git-provider-urls', '1.1.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),
        # Test with git provider configured in module provider and override in module version
        ('git-provider-urls', '1.4.0', 'https://base-url.com/repo_url_tests/git-provider-urls-test'),

        # Test with repo URLs configured in module provider
        ('module-provider-urls', '1.2.0', None),
        # Test with repo URls configured in module provider and override in module version
        ('module-provider-urls', '1.4.0', None),

        # Test with git provider configured in module provider and URLs configured in git provider
        ('module-provider-override-git-provider', '1.3.0', 'https://base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
        # Test with URls configured in module provider and override in module version
        ('module-provider-override-git-provider', '1.4.0', 'https://base-url.com/repo_url_tests/module-provider-override-git-provider-test'),
    ])
    def test_get_source_base_url_with_custom_module_provider_and_module_version_urls_disabled(self, module_name, module_version, expected_base_url):
        """Ensure base URL matches the expected values when module provider and module version repo urls are disabled."""
        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name=module_name)
        module_provider = ModuleProvider.get(module=module, name='test')
        assert module_provider is not None
        module_version = ModuleVersion.get(module_provider=module_provider, version=module_version)
        assert module_version is not None

        with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', False):
            with mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', False):
                assert module_version.get_source_base_url() == expected_base_url

    def test_get_total_count(self):
        """Test get_total_count method"""
        assert ModuleProvider.get_total_count() == 50

    def test_get_module_provider_existing(self):
        """Attempt to get existing module provider"""
        namespace = Namespace(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        module_provider = ModuleProvider.get(module=module, name='providername')
        assert module_provider is not None
        row = module_provider._get_db_row()
        assert row['id'] == 48
        assert row['namespace'] == 'genericmodules'
        assert row['module'] == 'modulename'
        assert row['provider'] == 'providername'

    def test_get_module_provider_non_existent(self):
        """Attempt to get non-existent module provider"""
        namespace = Namespace(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        module_provider = ModuleProvider.get(module=module, name='doesnotexist')
        assert module_provider is None

    def test_get_module_provider_with_create(self):
        """Attempt to get non-existent module provider with create"""
        namespace = Namespace(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', True):
            module_provider = ModuleProvider.get(module=module, name='doesnotexistgetcreate', create=True)
            assert module_provider is not None
            assert module_provider._get_db_row()['provider'] == 'doesnotexistgetcreate'

    def test_get_module_provider_with_create_auto_create_disabled(self):
        """Attempt to get non-existent module provider with auto-creation disabled"""
        namespace = Namespace(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False):
            module_provider = ModuleProvider.get(module=module, name='doesnotexist', create=True)
            assert module_provider is None

    def test_get_module_provider_with_create_existing(self):
        """Attempt to get non-existent module provider with create"""
        namespace = Namespace(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', True):
            module_provider = ModuleProvider.get(module=module, name='providername', create=True)
            assert module_provider is not None
            assert module_provider._get_db_row()['id'] == 48

    def test_get_module_provider_with_create_auto_create_disabled_existing(self):
        """Attempt to get non-existent module provider with auto-creation disabled"""
        namespace = Namespace(name='genericmodules')
        module = Module(namespace=namespace, name='modulename')
        with mock.patch('terrareg.config.Config.AUTO_CREATE_MODULE_PROVIDER', False):
            module_provider = ModuleProvider.get(module=module, name='providername', create=True)
            assert module_provider is not None
            assert module_provider._get_db_row()['id'] == 48
