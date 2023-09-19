
import os
from tempfile import mktemp
from unittest import mock
from urllib.parse import urlencode

import pytest
from pyop.exceptions import InvalidAuthenticationRequest, InvalidClientAuthentication, OAuthError
from oic.oic.message import AuthorizationResponse, AccessTokenResponse
from oic.oic.message import AuthorizationRequest
from werkzeug.datastructures import EnvironHeaders

from test.unit.terrareg import TerraregUnitTest
from test import client


@pytest.fixture
def mock_terraform_signing_key(request):
    """Create fake OIDC signing key"""
    # Signing RSA key
    signing_rsa_key = """
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDg9lttk9fpB7+PxpjVfZZPUC0NT8VGzzaT2qJlbyafY7HNPyBr
ixGc/EZbwx73FYhFnGW0IQd8xxTqlBZOFoAbI9Kx850J1J+gGn3IUbW3dm9aQq0d
cwMuhrMj45Ixiwd14cyGb+ZFsmGpdqRAEM2nbeQEnA5eNre0/uVGNuR+CQIDAQAB
AoGAdmk2NrdbLo2lh0hBqh4wwA6zqA4VCPCJCcpLMJkQ+1S+ggp4RiMtYjRn1GUg
J25uDDYGUooQJt2jZNYN54xwYNwXobGaCSlmWSfGfiCF6SKlVICf+d8EEYa8GcAM
rBDyTMghayn0oA03loSdAG5iqzF1ob/zQXgNCPJkc2C/IAECQQDwWRK2gt12edPh
kYr8XD9Hakjs8EaNEB4xO8GKCmnLhjRZDvMj5usXGkSfPo24qutssyYpn/nP6YR0
1/Q0mcNRAkEA75zI91DU82fMHhct2GgfEP2IvdaHHQ8zZnarC9Prn+6/6cNefhtN
S0+tiZj0R0B3dkLGTTqcmYSQe/EEjY2xOQJBAJnR9+b0s/W6HH91nUTLaPg0rn1t
fUmUci5CNyg4Z+MIfgItTjDA/d4oQpjD+QGh6dAEi70CFGga5Fm/SBxN+DECQBBV
7A2QYTRG+0+B3QpH7vZFkrD+ky+T/bkalga0Z/f7WvIg86w9SEO+JuKenujMqFhT
rRlOyaZdt0v73oeYBWECQQDc7n98Cx6G1Nt2/87o6UaYzW5N4SfWCPTaiS9/inpQ
yzEmVAlL/QfgkKm+0zsa8czkSwNjtBz9vOIffCxtZmlf
-----END RSA PRIVATE KEY-----
""".strip()
    signing_key_path = mktemp()
    with open(signing_key_path, "w") as signing_key_fh:
        signing_key_fh.write(signing_rsa_key)

    with mock.patch("terrareg.config.Config.TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH", signing_key_path):
        yield

    os.unlink(signing_key_path)


class TestTerraformOauth(TerraregUnitTest):
    """Test TerraformOauth endpoints class"""

    def test_authorization_endpoint_unhandled_exception(self, client, mock_terraform_signing_key):
        """Handle unknown error during parse_authentication_request"""

        mock_provider = mock.MagicMock()

        def raise_exception(*args, **kwargs):
            raise InvalidAuthenticationRequest(message="unittest error", parsed_request={}, oauth_error=None)
        mock_provider.parse_authentication_request.side_effect = raise_exception

        with mock.patch('terrareg.terraform_idp.TerraformIdp.provider', mock_provider):
            res = client.get('/terraform/oauth/authorization')

        assert res.status_code == 400
        assert res.text == "Something went wrong: unittest error"

    def test_authorization_endpoint_error(self, client, mock_terraform_signing_key):
        """Handle oauth error during parse_authentication_request"""

        mock_provider = mock.MagicMock()

        def raise_exception(*args, **kwargs):
            raise InvalidAuthenticationRequest(
                message="unittest error",
                parsed_request={
                    'client_id': 'terraform-cli',
                    'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
                    'code_challenge_method': 'S256',
                    'redirect_uri': 'http://localhost:10004/login',
                    'response_type': 'code',
                    'state': '2a01dd56-5b85-3248-b240-782085864837',
                    'scope': 'openid'
                },
                oauth_error="unauthorized_client"
            )
        mock_provider.parse_authentication_request.side_effect = raise_exception

        with mock.patch('terrareg.terraform_idp.TerraformIdp.provider', mock_provider):
            res = client.get('/terraform/oauth/authorization')

        assert res.status_code == 303
        assert res.headers['Location'] == "http://localhost:10004/login#error=unauthorized_client&error_message=unittest+error&state=2a01dd56-5b85-3248-b240-782085864837"

    def test_authorization_endpoint_authenticated(self, client, mock_terraform_signing_key):
        """Test authorization endpoint whilst authenticated"""

        authorization_request = AuthorizationRequest().deserialize(urlencode({
            'client_id': 'terraform-cli',
            'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
            'code_challenge_method': 'S256',
            'redirect_uri': 'http://localhost:10004/login',
            'response_type': 'code',
            'state': '2a01dd56-5b85-3248-b240-782085864837',
            'scope': 'openid'
        }))
        mock_parse_authentication_request = mock.MagicMock(return_value=authorization_request)

        authorize_response = AuthorizationResponse()
        authorize_response['code'] = '8647663211c1464d8454786887d01975'
        authorize_response['state'] = '2a01dd56-5b85-3248-b240-782085864837'

        mock_authorize = mock.MagicMock(return_value=authorize_response)

        mock_current_auth_method = mock.MagicMock()
        mock_current_auth_method.is_authenticated.return_value = True
        mock_current_auth_method.get_username.return_value = "Unittest username"

        with mock.patch('terrareg.terraform_idp.Provider.authorize', mock_authorize), \
                mock.patch('terrareg.terraform_idp.Provider.parse_authentication_request', mock_parse_authentication_request), \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_current_auth_method)), \
                mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example.local'), \
                mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT', 'supersecret'), \
                mock.patch('terrareg.config.Config.SECRET_KEY', 'supersecret'):

            # Update real app secret key
            self.SERVER._app.secret_key = 'averysecretkey'

            res = client.get('/terraform/oauth/authorization', query_string={
                'client_id': 'terraform-cli',
                'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
                'code_challenge_method': 'S256',
                'redirect_uri': 'http://localhost:10004/login',
                'response_type': 'code',
                'state': '2a01dd56-5b85-3248-b240-782085864837'
            })

        assert res.status_code == 303
        assert res.headers['Location'] == "http://localhost:10004/login?code=8647663211c1464d8454786887d01975&state=2a01dd56-5b85-3248-b240-782085864837"

        with client.session_transaction() as session:
            assert session == {
                'authn_req': {
                    'client_id': 'terraform-cli',
                    'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
                    'code_challenge_method': 'S256',
                    'redirect_uri': 'http://localhost:10004/login',
                    'response_type': 'code',
                    'scope': 'openid',
                    'state': '2a01dd56-5b85-3248-b240-782085864837'
                }
            }

        mock_parse_authentication_request.assert_called_once_with(
            'client_id=terraform-cli&code_challenge=IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM&code_challenge_method=S256&redirect_uri=http%3A%2F%2Flocalhost%3A10004%2Flogin&response_type=code&state=2a01dd56-5b85-3248-b240-782085864837&scope=openid',
            # Request headers - these aren't used,
            # and difficult to mock EnvironHeaders
            mock.ANY
        )
        mock_authorize.assert_called_once_with(authorization_request, "Unittest username")

    def test_authorization_endpoint_unauthenticated(self, client, mock_terraform_signing_key):
        """Test authorization endpoint whilst not authenticated"""

        authorization_request = AuthorizationRequest().deserialize(urlencode({
            'client_id': 'terraform-cli',
            'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
            'code_challenge_method': 'S256',
            'redirect_uri': 'http://localhost:10004/login',
            'response_type': 'code',
            'state': '2a01dd56-5b85-3248-b240-782085864837',
            'scope': 'openid'
        }))
        mock_parse_authentication_request = mock.MagicMock(return_value=authorization_request)

        authorize_response = AuthorizationResponse()
        authorize_response['code'] = '8647663211c1464d8454786887d01975'
        authorize_response['state'] = '2a01dd56-5b85-3248-b240-782085864837'

        mock_authorize = mock.MagicMock(return_value=authorize_response)

        mock_current_auth_method = mock.MagicMock()
        mock_current_auth_method.is_authenticated.return_value = False

        with mock.patch('terrareg.terraform_idp.Provider.authorize', mock_authorize), \
                mock.patch('terrareg.terraform_idp.Provider.parse_authentication_request', mock_parse_authentication_request), \
                mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_current_auth_method)), \
                mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example.local'), \
                mock.patch('terrareg.config.Config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT', 'supersecret'), \
                mock.patch('terrareg.config.Config.SECRET_KEY', 'supersecret'):

            # Update real app secret key
            self.SERVER._app.secret_key = 'averysecretkey'

            res = client.get('/terraform/oauth/authorization', query_string={
                'client_id': 'terraform-cli',
                'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
                'code_challenge_method': 'S256',
                'redirect_uri': 'http://localhost:10004/login',
                'response_type': 'code',
                'state': '2a01dd56-5b85-3248-b240-782085864837'
            })

        # Ensure user is redirected to login
        assert res.status_code == 302
        assert res.headers['Location'] == "/login?redirect=%2Fterraform%2Foauth%2Fauthorization%3Fclient_id%3Dterraform-cli%26code_challenge%3DIzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM%26code_challenge_method%3DS256%26redirect_uri%3Dhttp%253A%252F%252Flocalhost%253A10004%252Flogin%26response_type%3Dcode%26state%3D2a01dd56-5b85-3248-b240-782085864837"

        with client.session_transaction() as session:
            assert session == {
                'authn_req': {
                    'client_id': 'terraform-cli',
                    'code_challenge': 'IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM',
                    'code_challenge_method': 'S256',
                    'redirect_uri': 'http://localhost:10004/login',
                    'response_type': 'code',
                    'scope': 'openid',
                    'state': '2a01dd56-5b85-3248-b240-782085864837'
                }
            }

        mock_parse_authentication_request.assert_called_once_with(
            'client_id=terraform-cli&code_challenge=IzXlqzixEdheMSDy_biAVDFsycPZTLT83pgdLjDZZwM&code_challenge_method=S256&redirect_uri=http%3A%2F%2Flocalhost%3A10004%2Flogin&response_type=code&state=2a01dd56-5b85-3248-b240-782085864837&scope=openid',
            # Request headers - these aren't used,
            # and difficult to mock EnvironHeaders
            mock.ANY
        )
        mock_authorize.assert_not_called()


    def test_jwks_uri(self, client, mock_terraform_signing_key):
        """Test calling JWKS endpoint"""
        with mock.patch('terrareg.terraform_idp.Provider.jwks', {'some': 'jwks'}):

            res = client.get('/terraform/oauth/jwks')

        assert res.json == {'some': 'jwks'}

    def test_token_endpoint(self, client, mock_terraform_signing_key):
        """Handle valid request to token endpoint"""

        token = AccessTokenResponse().from_dict({
            'access_token': '4a20625d389e4257a7b02a8b6d5cd4d9',
            'token_type': 'Bearer',
            'expires_in': 3600,
            'id_token': 'eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiAiaHR0cHM6Ly9sb2NhbC1kZXYuZG9jay5zdHVkaW8iLCAic3ViIjogIkJ1aWx0LWluIGFkbWluIiwgImF1ZCI6IFsidGVycmFmb3JtLWNsaSJdLCAiaWF0IjogMTY5NDkzNTIxMywgImV4cCI6IDE2OTQ5Mzg4MTMsICJhdF9oYXNoIjogInoyUnRvQTh4dHVETE10UFoyX3o3VVEifQ.LMDBJnGbLztV9JLxfaiZXfWf4pQrAoAFzxfXNjyFShaiUdXXSTZkV7dif4bpOYxQ5uubYq7AAFRR508JVqaAG5nHdnvgo8TfvARrEn2qtWOyLZaCQWk8S-uN1kIKgAZA6yVBTbWODc0AY3eF5OtWpIVS5NDdkVbg7zN2FB3_4mS9styQzHk6NLMt6i05CXINGL_jxVjzVbeKmZs-NsIQK2vV8BdjE5UavizfDDn4Bg5CrzP61gY2N36Fv8EGnRvAo7F4Tb-bO7Nzv_BoiETM8A-CzMTxtZhf9PJJrLBXEPRUSnEVu-xaOa67AWnAvTtEcxoqQAyFby9vlTQY53pxlqwfN4MjdPt5QMIO40JjBoALD-qNuCamjKbmGykI9_tgX1-YoYKi2iHoni7czxrppmCE_O32hYToSHjZ3nYddP7SpFAN2m4jnkAjiKTPnenUP7QEEUoG1Mrv_sTXup6gbhl6Aqkz9COLRk79OLjTiG0XDQFGRCXzjkYO0_lqMZ-O5XHCIjpYoyYfyuZb0akziXA0CVaWUYiTBIXWH3MaIGmoRBigaJtHj43HO5LKX0vcqOmytOlowOVOdeJ9o279vnAmZBHBTA2yFLmqfKKrd9yW_K_83crF4UdfydxTQn90CkeT7oWYrV537mE4CqcgqthPF9xcPbW7fU1VVmCkNEc'
        })
        mock_handle_token_request = mock.MagicMock(return_value=token)

        with mock.patch('terrareg.terraform_idp.Provider.handle_token_request', mock_handle_token_request):
            res = client.post(
                "/terraform/oauth/token",
                data=b"client_id=terraform-cli&code=fdc4759847fc4bee935b2de812a96e12&code_verifier=5478dad3-af07-0753-0b37-3631f2810f1b.231116920&grant_type=authorization_code&redirect_uri=http%3A%2F%2Flocalhost%3A10005%2Flogin",
                headers={
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            )

        assert res.status_code == 200
        assert res.json == {
            'access_token': '4a20625d389e4257a7b02a8b6d5cd4d9',
            'expires_in': 3600,
            'id_token': 'eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiAiaHR0cHM6Ly9sb2NhbC1kZXYuZG9jay5zdHVkaW8iLCAic3ViIjogIkJ1aWx0LWluIGFkbWluIiwgImF1ZCI6IFsidGVycmFmb3JtLWNsaSJdLCAiaWF0IjogMTY5NDkzNTIxMywgImV4cCI6IDE2OTQ5Mzg4MTMsICJhdF9oYXNoIjogInoyUnRvQTh4dHVETE10UFoyX3o3VVEifQ.LMDBJnGbLztV9JLxfaiZXfWf4pQrAoAFzxfXNjyFShaiUdXXSTZkV7dif4bpOYxQ5uubYq7AAFRR508JVqaAG5nHdnvgo8TfvARrEn2qtWOyLZaCQWk8S-uN1kIKgAZA6yVBTbWODc0AY3eF5OtWpIVS5NDdkVbg7zN2FB3_4mS9styQzHk6NLMt6i05CXINGL_jxVjzVbeKmZs-NsIQK2vV8BdjE5UavizfDDn4Bg5CrzP61gY2N36Fv8EGnRvAo7F4Tb-bO7Nzv_BoiETM8A-CzMTxtZhf9PJJrLBXEPRUSnEVu-xaOa67AWnAvTtEcxoqQAyFby9vlTQY53pxlqwfN4MjdPt5QMIO40JjBoALD-qNuCamjKbmGykI9_tgX1-YoYKi2iHoni7czxrppmCE_O32hYToSHjZ3nYddP7SpFAN2m4jnkAjiKTPnenUP7QEEUoG1Mrv_sTXup6gbhl6Aqkz9COLRk79OLjTiG0XDQFGRCXzjkYO0_lqMZ-O5XHCIjpYoyYfyuZb0akziXA0CVaWUYiTBIXWH3MaIGmoRBigaJtHj43HO5LKX0vcqOmytOlowOVOdeJ9o279vnAmZBHBTA2yFLmqfKKrd9yW_K_83crF4UdfydxTQn90CkeT7oWYrV537mE4CqcgqthPF9xcPbW7fU1VVmCkNEc',
            'token_type': 'Bearer'
        }

        mock_handle_token_request.assert_called_once_with(
            "client_id=terraform-cli&code=fdc4759847fc4bee935b2de812a96e12&code_verifier=5478dad3-af07-0753-0b37-3631f2810f1b.231116920&grant_type=authorization_code&redirect_uri=http%3A%2F%2Flocalhost%3A10005%2Flogin",
            mock.ANY
        )

    def test_token_endpoint_invalid_client_auth(self, client, mock_terraform_signing_key):
        """Handle client auth error in token endpoint"""

        def raise_exception(*args, **kwargs):
            raise InvalidClientAuthentication("Unittest Invalid client authentication")
        mock_handle_token_request = mock.MagicMock(side_effect=raise_exception)

        with mock.patch('terrareg.terraform_idp.Provider.handle_token_request', mock_handle_token_request):
            res = client.post(
                "/terraform/oauth/token",
                data=b"client_id=terraform-cli&code=fdc4759847fc4bee935b2de812a96e12&code_verifier=5478dad3-af07-0753-0b37-3631f2810f1b.231116920&grant_type=authorization_code&redirect_uri=http%3A%2F%2Flocalhost%3A10005%2Flogin",
                headers={
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            )

        assert res.status_code == 401
        assert res.headers['Content-Type'] == 'application/json'
        assert res.headers['WWW-Authenticate'] == 'Basic'
        assert res.json == {
            'error': 'invalid_client',
            'error_description': 'Unittest Invalid client authentication',
        }

    def test_token_endpoint_invalid_oauth(self, client, mock_terraform_signing_key):
        """Handle OauthError in token endpoint"""

        def raise_exception(*args, **kwargs):
            raise OAuthError(message="Unit test oauth error", oauth_error="unittest_invalid")
        mock_handle_token_request = mock.MagicMock(side_effect=raise_exception)

        with mock.patch('terrareg.terraform_idp.Provider.handle_token_request', mock_handle_token_request):
            res = client.post(
                "/terraform/oauth/token",
                data=b"client_id=terraform-cli&code=fdc4759847fc4bee935b2de812a96e12&code_verifier=5478dad3-af07-0753-0b37-3631f2810f1b.231116920&grant_type=authorization_code&redirect_uri=http%3A%2F%2Flocalhost%3A10005%2Flogin",
                headers={
                    "Content-Type": "application/x-www-form-urlencoded"
                }
            )

        assert res.status_code == 400
        assert res.headers['Content-Type'] == 'application/json'
        assert res.json == {
            'error': 'unittest_invalid',
            'error_description': 'Unit test oauth error',
        }
