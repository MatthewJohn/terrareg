
from test.unit.terrareg import TerraregUnitTest
from test import client


class TestTerraformWellKnown(TerraregUnitTest):
    """Test TerraformWellKnown resource."""

    def test_with_no_params(self, client):
        """Test endpoint without query parameters"""
        res = client.get('/.well-known/terraform.json')
        assert res.status_code == 200
        assert res.json == {
            'modules.v1': '/v1/modules/'
        }

    def test_with_post(self, client):
        """Test invalid post request"""
        res = client.post('/.well-known/terraform.json')
        assert res.status_code == 405
