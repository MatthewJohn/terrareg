
import datetime
from unittest import mock

import pytest

import terrareg.errors
from test.unit.terrareg import TerraregUnitTest
from test import client, app_context, test_request_context
import terrareg.auth


class TestGithubAuthCallback(TerraregUnitTest):
    """Test Github Auth callback API."""

    def test_without_code(self):
        raise NotImplementedError

    def test_unable_to_get_access_token(self):
        raise NotImplementedError

    def test_unable_to_get_username(self):
        raise NotImplementedError

    def test_call(self):
        raise NotImplementedError
