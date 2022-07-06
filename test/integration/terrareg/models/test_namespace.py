
import pytest

from terrareg.models import Namespace
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest

class TestNamespace(TerraregIntegrationTest):

    @pytest.mark.parametrize('namespace_name', [
        'invalid@atsymbol',
        'invalid"doublequote',
        "invalid'singlequote",
        '-startwithdash',
        'endwithdash-',
        '_startwithunderscore',
        'endwithunscore_',
        'a:colon',
        'or;semicolon',
        'who?knows'
    ])
    def test_invalid_namespace_names(self, namespace_name):
        """Test invalid namespace names"""
        with pytest.raises(terrareg.errors.InvalidNamespaceNameError):
            Namespace(name=namespace_name)

    @pytest.mark.parametrize('namespace_name', [
        'normalname',
        'name2withnumber',
        '2startendiwthnumber2',
        'contains4number',
        'with-dash',
        'with_underscore',
        'withAcapital',
        'StartwithCaptital',
        'endwithcapitaL'
    ])
    def test_valid_namespace_names(self, namespace_name):
        """Test valid namespace names"""
        Namespace(name=namespace_name)

    def test_get_total_count(self):
        """Test get_total_count method"""
        assert Namespace.get_total_count() == 11
