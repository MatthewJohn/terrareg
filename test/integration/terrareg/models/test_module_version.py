
import pytest

from terrareg.models import Module, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest

class TestModuleProvider(TerraregIntegrationTest):

    @pytest.mark.parametrize('version', [
        'astring',
        '',
        '1',
        '1.1',
        '.23.1',
        '1.1.1.1',
        '1.1.1.',
        '1.2.3-a',
        '1.2.2-1'
    ])
    def test_invalid_module_versions(self, version):
        """Test invalid module versions"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        module_provider = ModuleProvider(module=module, name='test')
        with pytest.raises(terrareg.errors.InvalidVersionError):
            ModuleVersion(module_provider=module_provider, version=version)

    @pytest.mark.parametrize('version', [
        '1.1.1',
        '13.14.16',
        '1.10.10',
        '01.01.01'
    ])
    def test_valid_module_versions(self, version):
        """Test valid module versions"""
        namespace = Namespace(name='test')
        module = Module(namespace=namespace, name='test')
        module_provider = ModuleProvider(module=module, name='test')
        ModuleVersion(module_provider=module_provider, version=version)
