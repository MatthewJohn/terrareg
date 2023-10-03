
from flask import session, make_response, render_template, redirect

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.github


class GithubLoginInitiate(ErrorCatchingResource):
    """Interface to initiate authentication via Github"""

    def _get(self):
        """Redirect to github login."""

        if not terrareg.github.Github.is_enabled():
            return make_response(render_template(
                'error.html',
                error_title='Login error',
                error_description='Github authentication is not enabled'
            ))

        redirect_url = terrareg.github.Github.get_login_redirect_url()

        return redirect(redirect_url, code=302)
