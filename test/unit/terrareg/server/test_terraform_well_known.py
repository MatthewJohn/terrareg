
from unittest import mock
from test.unit.terrareg import TerraregUnitTest
from test import client


class TestTerraformWellKnown(TerraregUnitTest):
    """Test TerraformWellKnown resource."""

    def test_with_no_params(self, client):
        """Test endpoint without query parameters"""
        with mock.patch('terrareg.terraform_idp.TerraformIdp.is_enabled', False):
            res = client.get('/.well-known/terraform.json')
        assert res.status_code == 200
        assert res.json == {
            'modules.v1': '/v1/modules/'
        }

    def test_with_terraform_idc(self, client):
        """Test endpoint without query parameters"""
        with mock.patch('terrareg.terraform_idp.TerraformIdp.is_enabled', True):
            res = client.get('/.well-known/terraform.json')

        assert res.status_code == 200
        assert res.json == {
            'modules.v1': '/v1/modules/',
            'login.v1': {
                'authz': '/terraform/oauth/authorization',
                'client': 'terraform-cli',
                'grant_types': ['authz_code', 'token'],
                'ports': [10000, 10010],
                'token': '/terraform/oauth/token'
            }
        }

    def test_with_post(self, client):
        """Test invalid post request"""
        res = client.post('/.well-known/terraform.json')
        assert res.status_code == 405
