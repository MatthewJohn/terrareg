
from unittest import mock
import pytest
from terrareg.database import Database

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

    @pytest.mark.parametrize('display_name', [
        '< Is not valid',
        'Not valid !',
        ' ',
        'Is @ Not valid',
        'Character test (',
        'Character test )',
        'Character test *',
        'Character test &',
        'Character test ^',
        '-Start with dash',
        'End with dash-',
        '_Start with underscore',
        'End with underscore_',
        ' Start with space',
        'End with space ',
    ])
    def test_validate_display_name_invalid_name(self, display_name):
        """Test _validate_display_name with invalid name"""
        with pytest.raises(terrareg.errors.InvalidNamespaceDisplayNameError):
            Namespace._validate_display_name(display_name=display_name)

    @pytest.mark.parametrize('display_name', [
        None,
        '',
        'SingleWord',
        'With Spaces',
        'With_Underscore',
        'With-Dash',
        'WithNumb3r5'
    ])
    def test_validate_display_name_valid_name(self, display_name):
        """Test _validate_display_name with invalid name"""
        Namespace._validate_display_name(display_name=display_name)

    def test_validate_display_name_duplicate_name(self):
        """Test _validate_display_name with invalid name"""
        original = Namespace.create(name="test-duplicate", display_name="Test Duplicate")
        try:
            with pytest.raises(terrareg.errors.DuplicateNamespaceDisplayNameError):
                Namespace._validate_display_name("Test Duplicate")

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.id==original.pk))

    @pytest.mark.parametrize('name,display_name', [
        ('test-create-namespace', 'Test Create Namespace'),
        ('test-create-namespace', '')
    ])
    def test_create(self, name, display_name):
        """Test create method of namespace"""
        try:
            namespace = Namespace.create(name=name, display_name=display_name)

            assert namespace.pk
            assert namespace.name == name
            if display_name == "":
                assert namespace.display_name is None
            else:
                assert namespace.display_name == display_name

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace==name))

    def test_create_with_name_error(self):
        """Test Namespace create method with _validate_name error"""
        def raise_error(name):
            raise terrareg.errors.InvalidNamespaceNameError("Invalid Namespace name")

        with mock.patch("terrareg.models.Namespace._validate_name", mock.MagicMock(side_effect=raise_error)):
            with pytest.raises(terrareg.errors.InvalidNamespaceNameError):
                Namespace.create(name="test", display_name="")

    def test_create_with_display_name_error(self):
        """Test Namespace create method with _validate_display_name error"""
        def raise_error(display_name):
            raise terrareg.errors.InvalidNamespaceDisplayNameError("Invalid Namespace display name")

        with mock.patch("terrareg.models.Namespace._validate_display_name", mock.MagicMock(side_effect=raise_error)):
            with pytest.raises(terrareg.errors.InvalidNamespaceDisplayNameError):
                Namespace.create(name="test", display_name="")

    def test_create_duplicate(self):
        """Create creating namespace with duplicate name"""
        original = Namespace.create(name="test-duplicate", display_name="")
        try:
            with pytest.raises(terrareg.errors.NamespaceAlreadyExistsError):
                Namespace.create(name="test-duplicate", display_name="")

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.id==original.pk))

    def test_create_duplicate_empty_display_name(self):
        """Create creating two namespaces with empty display name"""
        Namespace.create(name="test-duplicate", display_name="")
        try:
            Namespace.create(name="test-duplicate2", display_name="")

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-duplicate"))
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-duplicate2"))

