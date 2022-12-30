
import re

from flask import request, make_response

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Namespace
import terrareg.config
from terrareg.analytics import AnalyticsEngine


class ApiModuleVersionDownload(ErrorCatchingResource):
    """Provide download endpoint."""

    def _get(self, namespace, name, provider, version):
        """Provide download header for location to download source."""
        namespace, analytics_token = Namespace.extract_analytics_token(namespace)
        namespace, module, module_provider, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return self._get_404_response()

        auth_token = None
        auth_token_match = re.match(r'Bearer (.*)', request.headers.get('Authorization', ''))
        if auth_token_match:
            auth_token = auth_token_match.group(1)

        # Determine if auth token is for internal initialisation of modules
        # during module extraction
        if auth_token == terrareg.config.Config().INTERNAL_EXTRACTION_ANALYITCS_TOKEN:
            pass
        # otherwise, if module download should be rejected due to
        # non-existent analytics token
        elif not analytics_token and not terrareg.config.Config().ALLOW_UNIDENTIFIED_DOWNLOADS:
            return make_response(
                ("\nAn {analytics_token_phrase} must be provided.\n"
                 "Please update module source to include {analytics_token_phrase}.\n"
                 "\nFor example:\n  source = \"{host}/{example_analytics_token}__{namespace}/{module_name}/{provider}\"").format(
                    analytics_token_phrase=terrareg.config.Config().ANALYTICS_TOKEN_PHRASE,
                    host=request.host,
                    example_analytics_token=terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN,
                    namespace=namespace.name,
                    module_name=module.name,
                    provider=module_provider.name
                ),
                401
            )
        else:
            # Otherwise, if download is allowed and not internal, record the download
            AnalyticsEngine.record_module_version_download(
                module_version=module_version,
                analytics_token=analytics_token,
                terraform_version=request.headers.get('X-Terraform-Version', None),
                user_agent=request.headers.get('User-Agent', None),
                auth_token=auth_token
            )

        resp = make_response('', 204)
        resp.headers['X-Terraform-Get'] = module_version.get_source_download_url()
        return resp

