
import datetime
from unittest import mock
import pytest

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.auth import AuthenticationType, GithubAuthMethod
from test import BaseTest
from test.unit.terrareg import setup_test_data, mock_models
from test.unit.terrareg.auth.base_session_auth_method_tests import BaseSessionAuthMethodTests
from test.unit.terrareg.auth.base_sso_auth_method_tests import BaseSsoAuthMethodTests, test_data, user_group_data

# Required as this is used by BaseOpenidConnectAuthMethod
from test import test_request_context


class TestGithubAuthMethod(BaseSsoAuthMethodTests, BaseSessionAuthMethodTests):
    """Test methods of Github auth method"""

    CLS = GithubAuthMethod

    def test__get_organisation_memeberships(self):
        """Test _get_organisation_memeberships"""
        raise NotImplementedError

    def test_get_group_memberships(self):
        """Test get_group_memberships"""
        raise NotImplementedError

    def test_get_group_memberships(self):
        """test get_group_memberships."""
        # If github automatic namespace generation is enabled,
        # allow access to these namespaces
        raise NotImplementedError

    def test_get_all_namespace_permissions(self):
        """test test_get_all_namespace_permissions"""
        raise NotImplementedError
