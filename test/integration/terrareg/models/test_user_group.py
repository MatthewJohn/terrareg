
import datetime
import secrets
from unittest import mock
import pytest
from terrareg.database import Database

from terrareg.models import Namespace, Session, UserGroup
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest

class TestUserGroup(TerraregIntegrationTest):
    """Test UserGroup model class"""

    def setup_method(self, method):
        """Remove any pre-existing user groups and permissions before running each test."""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.user_group_namespace_permission.delete())
            conn.execute(db.user_group.delete())
        super(TestUserGroup, self).setup_method(method)

    @pytest.mark.parametrize('user_group_name', [
        'test',
        'test_group', '_testgroup', 'testgroup_',
        'test-group', '-testgroup', 'testgroup-',
        'test group', ' testgroup', 'testgroup '
    ])
    def test_create_user_group(self, user_group_name):
        """Test create method, successfully creating user group"""
        user_group = UserGroup.create(user_group_name, site_admin=False)
        assert user_group.name == user_group_name
        assert user_group.site_admin == False

    @pytest.mark.parametrize('user_group_name', [
        'invalid@group', '@invalidgroup', 'invalidgroup@',
        'invalid#group', '#invalidgroup', 'invalidgroup#',
        'invalid"group', '"invalidgroup', 'invalidgroup"',
        "invalid'group", "'invalidgroup", "invalidgroup'",
    ])
    def test_create_user_group_invalid_name(self, user_group_name):
        """Test create method with an invalid group name"""
        with pytest.raises(terrareg.errors.InvalidUserGroupNameError):
            UserGroup.create(user_group_name, site_admin=False)
        assert UserGroup.get_all_user_groups() == []

    def test_create_duplicate_user_group_name(self):
        """Test creating a user group that has a duplicate name"""
        user_group = UserGroup.create('duplicate', site_admin=True)
        assert len(UserGroup.get_all_user_groups()) == 1
        user_group_pk = user_group.pk

        second_user_group = UserGroup.create('duplicate', site_admin=False)
        assert second_user_group == None

        # Check user group was not created and original user group
        # has been unchanged
        all_user_groups = UserGroup.get_all_user_groups()
        assert len(all_user_groups) == 1
        assert all_user_groups[0].site_admin == True
        assert all_user_groups[0].pk == user_group_pk

    @pytest.mark.parametrize('site_admin', [
        True,
        False
    ])
    def test_create_with_site_admin(self, site_admin):
        """Test creating user group with varying site_admin attribute values"""
        user_group = UserGroup.create('group', site_admin=site_admin)
        assert user_group.site_admin == site_admin
        assert UserGroup.get_all_user_groups()[0].site_admin == site_admin

    def test_get_by_group_name(self):
        """Test get_by_group_name method"""
        # Test without any user groups setup
        user_group = UserGroup.get_by_group_name('doesnotexist')
        assert user_group is None

        # Setup test groups
        test_group1_db_row = UserGroup.create('testgroup1', site_admin=True)._get_db_row()
        UserGroup.create('testgroup2', site_admin=True)
        UserGroup.create('testgroup3', site_admin=False)

        # Test getting group
        user_group = UserGroup.get_by_group_name('testgroup1')
        assert isinstance(user_group, UserGroup)
        assert user_group.name == 'testgroup1'
        assert user_group._get_db_row() == test_group1_db_row

        # Test getting non-existent group
        assert UserGroup.get_by_group_name('doesnotexist') == None

        # Ensure partial name match does not match
        assert UserGroup.get_by_group_name('testgroup') == None

    def test_get_all_user_groups(self):
        """Test get_all_user_groups method"""
        # Test without any user groups setup
        user_group = UserGroup.get_all_user_groups()
        assert user_group == []

        # Setup test groups
        test_group1 = UserGroup.create('testgroup1', site_admin=True)
        test_group2 = UserGroup.create('testgroup2', site_admin=True)
        test_group3 = UserGroup.create('testgroup3', site_admin=False)

        # Test getting group
        user_groups = UserGroup.get_all_user_groups()
        assert len(user_groups) == 3
        assert test_group1 in user_groups
        assert test_group2 in user_groups
        assert test_group3 in user_groups
