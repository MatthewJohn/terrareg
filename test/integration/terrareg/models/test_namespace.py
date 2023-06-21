
from unittest import mock
import pytest

from terrareg.audit import AuditEvent
from terrareg.database import Database
import terrareg.audit_action
from terrareg.models import Namespace
import terrareg.models
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
        'who?knows',
        'contains__doubleunderscore',
        '-a',
        'a-',
        'a_',
        '_a',
        '__',
        '--',
        '_',
        '-',
    ])
    def test_invalid_namespace_names(self, namespace_name):
        """Test invalid namespace names"""
        with pytest.raises(terrareg.errors.InvalidNamespaceNameError):
            Namespace._validate_name(name=namespace_name)

    @pytest.mark.parametrize('namespace_name', [
        'normalname',
        'name2withnumber',
        '2startendiwthnumber2',
        'contains4number',
        'with-dash',
        'with_underscore',
        'withAcapital',
        'StartwithCaptital',
        'endwithcapitaL',
        # Two letters
        'tl',
        # Two numbers
        '11',
        # Two characters with dash/unserscore
        'a-z',
        'a_z',
    ])
    def test_valid_namespace_names(self, namespace_name):
        """Test valid namespace names"""
        Namespace._validate_name(name=namespace_name)

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
                Namespace.create(name="is-not-duplicate", display_name="Test Duplicate")

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

    def test_update_display_name(self):
        """Test updating display name of namespace"""
        try:
            ns = Namespace.create(name="test-update-display-name", display_name="Old display name")

            # Update display name
            ns.update_display_name("New Display Name")

            check_ns = Namespace.get(name="test-update-display-name")
            assert check_ns.display_name == "New Display Name"

            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            audit_event = audit_events[0]
            assert audit_event['action'] == terrareg.audit_action.AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME
            assert audit_event['object_type'] == "Namespace"
            assert audit_event['object_id'] == "test-update-display-name"
            assert audit_event['old_value'] == "Old display name"
            assert audit_event['new_value'] == "New Display Name"

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-update-display-name"))

    def test_update_display_name_duplicate(self):
        """Test updating display name of namespace"""
        try:
            ns = Namespace.create(name="test-update-display-name", display_name="Old display name")
            Namespace.create(name="test-update-display-name-duplicate", display_name="Duplicate display name")

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            with pytest.raises(terrareg.errors.DuplicateNamespaceDisplayNameError):
                # Update display name
                ns.update_display_name("Duplicate display name")

            check_ns = Namespace.get(name="test-update-display-name")
            assert check_ns.display_name == "Old display name"

            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            assert len(audit_events) == 0

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-update-display-name"))
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-update-display-name-duplicate"))

    @pytest.mark.parametrize('old_value, new_value', [
        # Test same value
        ('Old display name', 'Old display name'),
        # Test various empty values
        (None, None),
        (None, ''),
        ('', None),
        ('', ''),
    ])
    def test_update_display_name_without_change(self, old_value, new_value):
        """Test updating display name of namespace with same name"""
        try:
            ns = Namespace.create(name="test-update-display-name", display_name=old_value)

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            # Update display name
            ns.update_display_name(new_value)

            check_ns = Namespace.get(name="test-update-display-name")
            # Check old value is still used (None is returned instead of empty strings)
            assert check_ns.display_name == (old_value or None)

            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            assert len(audit_events) == 0

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-update-display-name"))

    def test_update_display_name_capitalisation_change(self):
        """
        Test updating display name of namespace with same name with different capitalisation.

        This will invoke the display name change functionality (as the name is different),
        but will clash with itself when checking for duplicates.
        """
        try:
            ns = Namespace.create(name="test-update-display-name", display_name="Old display name")

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            # Update display name
            ns.update_display_name("Old Display NAME")

            check_ns = Namespace.get(name="test-update-display-name")
            assert check_ns.display_name == "Old Display NAME"

            # Check audit event
            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            audit_event = audit_events[0]
            assert audit_event['action'] == terrareg.audit_action.AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME
            assert audit_event['object_type'] == "Namespace"
            assert audit_event['object_id'] == "test-update-display-name"
            assert audit_event['old_value'] == "Old display name"
            assert audit_event['new_value'] == "Old Display NAME"

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-update-display-name"))

    def test_update_name(self):
        """Test updating name of namespace"""
        try:
            ns = Namespace.create(name="test-change-name")
            old_pk = ns.pk

            # Remove all namespace redirects
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace_redirect.delete())

            # Update display name
            ns.update_name("new-changed-name")

            check_ns = Namespace.get(name="new-changed-name", include_redirect=False)
            assert check_ns is not None
            assert check_ns.name == "new-changed-name"
            assert check_ns.pk == old_pk

            # Ensure a namespace redirect has been setup with the old name
            namespace_from_redirect = terrareg.models.NamespaceRedirect.get_namespace_by_name("test-change-name")
            assert namespace_from_redirect is not None
            assert namespace_from_redirect.pk == old_pk

            # Ensure the namespace can only be obtained using redirect
            assert Namespace.get(name="test-change-name", include_redirect=False) is None
            namespace_from_redirect = Namespace.get(name="test-change-name", include_redirect=True)
            assert namespace_from_redirect is not None
            assert namespace_from_redirect.pk == old_pk

            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            audit_event = audit_events[0]
            assert audit_event['action'] == terrareg.audit_action.AuditAction.NAMESPACE_MODIFY_NAME
            assert audit_event['object_type'] == "Namespace"
            assert audit_event['object_id'] == "test-change-name"
            assert audit_event['old_value'] == "test-change-name"
            assert audit_event['new_value'] == "new-changed-name"

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-change-name"))
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="new-changed-name"))

    def test_update_name_duplicate(self):
        """Test updating name of namespace with a duplicate name"""
        try:
            ns = Namespace.create(name="test-change-name")
            duplicate_ns = Namespace.create(name="test-change-name-duplicate")
            old_ns_pk = ns.pk
            duplicate_ns_pk = duplicate_ns.pk

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            # Remove all namespace redirects
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace_redirect.delete())

            with pytest.raises(terrareg.errors.NamespaceAlreadyExistsError):
                # Update display name
                ns.update_name("test-change-name-duplicate")

            check_ns = Namespace.get(name="test-change-name")
            assert check_ns is not None
            assert check_ns.pk == old_ns_pk

            check_duplicate_ns = Namespace.get(name="test-change-name-duplicate")
            assert check_duplicate_ns is not None
            assert check_duplicate_ns.pk == duplicate_ns_pk

            # Ensure no Namespace redirect was created
            with db.get_connection() as conn:
                assert len(conn.execute(db.namespace_redirect.select()).all()) == 0

            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            assert len(audit_events) == 0

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-change-name"))
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-change-name-duplicate"))

    def test_update_name_without_change(self):
        """Test updating name of namespace with same name"""
        try:
            ns = Namespace.create(name="test-update-name")
            old_pk = ns.pk

            # Remove all audit events
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.audit_history.delete())

            # Remove all namespace redirects
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace_redirect.delete())

            # Update display name
            ns.update_name("test-update-name")

            check_ns = Namespace.get(name="test-update-name")
            assert check_ns is not None
            assert check_ns.pk == old_pk

            # Check audit event
            audit_events, _, _ = AuditEvent.get_events(limit=1, descending=True, order_by="timestamp")
            assert len(audit_events) == 0

            # Ensure no Namespace redirect was created
            with db.get_connection() as conn:
                assert len(conn.execute(db.namespace_redirect.select()).all()) == 0

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="test-update-name"))
