
import datetime
from unittest import mock

import pytest

import terrareg.errors
from test.unit.terrareg import TerraregUnitTest
from test import client, app_context, test_request_context
import terrareg.auth


class TestGithubAuthInitiate(TerraregUnitTest):
    """Test Github Auth initiate API."""

    def test_without_sessions_enabled(self):
        raise NotImplementedError

    def test_without_github_configurations(self):
        raise NotImplementedError

    def test_call(self):
        raise NotImplementedError
