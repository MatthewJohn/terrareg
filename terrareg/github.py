
from urllib.parse import parse_qs

import requests

import terrareg.config


class Github:

    @classmethod
    def is_enabled(self):
        """Whether github authentication is enabled"""
        config = terrareg.config.Config()
        return bool(config.GITHUB_APP_CLIENT_ID and config.GITHUB_APP_CLIENT_SECRET and config.GITHUB_URL)

    @classmethod
    def get_login_redirect_url(cls):
        """Generate login redirect URL"""
        config = terrareg.config.Config()
        return f"{config.GITHUB_URL}/login/oauth/authorize?client_id={config.GITHUB_APP_CLIENT_ID}"

    @classmethod
    def get_access_token(cls, code):
        """Obtain access token from code"""
        config = terrareg.config.Config()
        res = requests.post(
            f"{config.GITHUB_URL}/login/oauth/access_token",
            data={
                "client_id": config.GITHUB_APP_CLIENT_ID,
                "client_secret": config.GITHUB_APP_CLIENT_SECRET,
                "code": code
            }
        )
        if res.status_code == 200:
            data = parse_qs(res.text)
            if (access_tokens := data.get("access_token")) and len(access_tokens) == 1:
                return access_tokens[0]
