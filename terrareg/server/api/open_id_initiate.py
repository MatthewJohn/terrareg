
from flask import session, make_response, render_template, redirect

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.openid_connect


class ApiOpenIdInitiate(ErrorCatchingResource):
    """Interface to initiate authentication via OpenID connect"""

    def _get(self):
        """Generate session for storing OpenID state token and redirect to openid login provider."""
        redirect_url, state = terrareg.openid_connect.OpenidConnect.get_authorize_redirect_url()

        if redirect_url is None:
            res = make_response(render_template(
                'error.html',
                error_title='Login error',
                error_description='SSO is incorrectly configured'
            ))
            res.headers['Content-Type'] = 'text/html'
            return res

        session['openid_connect_state'] = state
        session.modified = True

        return redirect(redirect_url, code=302)