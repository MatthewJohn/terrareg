
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
