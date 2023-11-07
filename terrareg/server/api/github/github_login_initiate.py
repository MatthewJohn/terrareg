
from flask import session, make_response, render_template, redirect

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.github
import terrareg.provider_source.factory


class GithubLoginInitiate(ErrorCatchingResource):
    """Interface to initiate authentication via Github"""

    def _get(self, provider_source):
        """Redirect to github login."""
        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if not provider_source_obj:
            return make_response(self._render_template(
                'error.html',
                error_title='Login error',
                error_description=f'{provider_source} authentication is not enabled',
                root_bread_brumb='Login'
            ))


        if self.create_session() is None:
            return make_response(render_template(
                'error.html',
                error_title='Login error',
                error_description='Sessions are not available'
            ))

        redirect_url = provider_source_obj.get_login_redirect_url()

        # Set session for provider source
        session['provider_source'] = provider_source_obj.name

        return redirect(redirect_url, code=302)
