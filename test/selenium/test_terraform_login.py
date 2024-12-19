
import base64
import hashlib
import os
import random
import socketserver
import string
from tempfile import NamedTemporaryFile
import threading
from time import sleep
from unittest import mock
import http.server
from urllib.parse import parse_qs, urlparse

from selenium.webdriver.common.by import By
import requests

from test.selenium import SeleniumTest



class HandleTestRequestHandler(http.server.BaseHTTPRequestHandler):

    requests = []

    def do_GET(self, post_vars={}):
        HandleTestRequestHandler.requests.append(self.path)
        self.send_response(200, "OK")
        self.send_header('Content-type', "text/plain")
        self.end_headers()
        self.wfile.write(b"Hit Unittest Server")


class TestTerraformLogin(SeleniumTest):
    """Test terraform login"""

    _SECRET_KEY = "354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe"

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""

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
        cls.signing_key_path = NamedTemporaryFile().name
        with open(cls.signing_key_path, "w") as signing_key_fh:
            signing_key_fh.write(signing_rsa_key)

        cls.register_patch(mock.patch("terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN", "unittest-password"))
        cls.register_patch(mock.patch("terrareg.config.Config.TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT", "super-secret"))
        cls.register_patch(mock.patch("terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS", False))
        cls.register_patch(mock.patch("terrareg.config.Config.DEBUG", True))
        cls.register_patch(mock.patch("terrareg.config.Config.TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH", cls.signing_key_path))
        # Mock this - It's not currently used by the IDP to determine any URLs, so
        # doesn't need to be correct
        cls.register_patch(mock.patch("terrareg.config.Config.PUBLIC_URL", "https://localhost"))

        super(TestTerraformLogin, cls).setup_class()

    @classmethod
    def teardown_class(cls):
        """Clear down any cookes from the trst."""

        if os.path.isfile(cls.signing_key_path):
            os.unlink(cls.signing_key_path)

        super(TestTerraformLogin, cls).teardown_class()

    def test_terraform_login(self):
        """Test Terraform login"""

        terraform_wellknown = requests.get(self.get_url("/.well-known/terraform.json")).json()

        # Create fake Terraform login server
        run_test_server = True
        terraform_server_port = 10000
        def start_server():
            for port in range(10000, 10005):
                nonlocal terraform_server_port
                terraform_server_port = port
                try:
                    with socketserver.TCPServer(("", port), HandleTestRequestHandler) as httpd:
                        while run_test_server:
                            httpd.handle_request()

                    # Once server has successfully finished,
                    # don't try any more ports
                    break
                except OSError:
                    pass


        web_server_thread = threading.Thread(target=start_server)
        web_server_thread.start()

        # Wait for server to start handling requests
        for _ in range(10):
            try:
                requests.get(f"http://localhost:{terraform_server_port}/ping")
                break
            except:
                # Server not started yet
                sleep(0.5)
        else:
            raise

        try:
            # Code verifier and challenge
            code_verifier = "areallysecretcodeverifier"            
            code_challenge = hashlib.sha256(code_verifier.encode('utf-8')).digest()
            code_challenge = base64.urlsafe_b64encode(code_challenge).decode('utf-8')
            code_challenge = code_challenge.replace('=', '')

            self.selenium_instance.get(
                self.get_url(
                    f"{terraform_wellknown['login.v1']['authz']}?"
                    f"client_id={terraform_wellknown['login.v1']['client']}&"
                    f"code_challenge={code_challenge}&"
                    "code_challenge_method=S256&"
                    f"redirect_uri=http%3A%2F%2Flocalhost%3A{terraform_server_port}%2Flogin&"
                    "response_type=code&"
                    "state=8cf3ee58-8c5d-5d45-475a-0a56e3d00aac"
                )
            )

            # Ensure user is redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url.startswith(self.get_url("/login?redirect=")), True)

            # Login using form
            token_input_field = self.selenium_instance.find_element(By.ID, 'admin_token_input')
            token_input_field.send_keys("unittest-password")
            login_button = self.selenium_instance.find_element(By.ID, 'login-button')
            login_button.click()

            # Ensure user is redirected to Terraform redirect URL
            self.assert_equals(lambda: self.selenium_instance.current_url.startswith(f'http://localhost:{terraform_server_port}'), True)
            redirect_url = self.selenium_instance.current_url
            query_string_args = parse_qs(urlparse(redirect_url).query)

            # Ensure state has not changed
            assert query_string_args["state"][0] == "8cf3ee58-8c5d-5d45-475a-0a56e3d00aac"

            # Obtain code
            code = query_string_args["code"][0]

            # Call to token endpoint
            token_res = requests.post(
                self.get_url(terraform_wellknown["login.v1"]["token"]),
                headers={
                    "Content-Type": "application/x-www-form-urlencoded"
                },
                data=(
                    "client_id=terraform-cli&"
                    f"code={code}&"
                    f"code_verifier={code_verifier}&"
                    "grant_type=authorization_code&"
                    f"redirect_uri=http%3A%2F%2Flocalhost%3A{terraform_server_port}%2Flogin"
                )
            )
            assert token_res.status_code == 200

            # Ensure access token is present in response
            token_res_json = token_res.json()
            assert "access_token" in token_res_json
            access_token = token_res_json["access_token"]

            # Attempt to access resource with token
            resource_res = requests.get(
                self.get_url("/v1/modules/adgadg__moduledetails/fullypopulated/testprovider/1.5.0/download"),
                headers={"Authorization": f"Bearer {access_token}"}
            )
            assert resource_res.status_code == 204

        finally:

            run_test_server = False
            requests.get(f"http://localhost:{terraform_server_port}/stop")
            web_server_thread.join()
