
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
        if not code:
            return None

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

    @classmethod
    def get_username(cls, access_token):
        """Get username of authenticated user"""
        if not access_token:
            return None
        config = terrareg.config.Config()
        res = requests.get(
            f"{config.GITHUB_API_URL}/user",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {access_token}"
            }
        )
        if res.status_code == 200:
            return res.json().get("login")

    @classmethod
    def get_user_organisations(cls, access_token):
        """Get username of authenticated user"""
        if not access_token:
            return []

        config = terrareg.config.Config()
        res = requests.get(
            f"{config.GITHUB_API_URL}/user/memberships/orgs",
            headers={
                "X-GitHub-Api-Version": "2022-11-28",
                "Accept": "application/vnd.github+json",
                "Authorization": f"Bearer {access_token}"
            }
        )

        if res.status_code == 200 and (response_data := res.json()):
            # Iterate over memberships, only get active memberships
            # that where the user is admin
            return [
                org_membership.get("organization", {}).get("login")
                for org_membership in response_data
                if (
                    org_membership.get("organization", {}).get("login") and
                    org_membership.get("state") == "active" and
                    org_membership.get("role") == "admin"
                )
            ]
        return []
