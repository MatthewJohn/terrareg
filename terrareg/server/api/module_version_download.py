
import re

from flask import request, make_response

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.config
import terrareg.analytics


class ApiModuleVersionDownload(ErrorCatchingResource):
    """Provide download endpoint."""

    def _get(self, namespace, name, provider, version=None):
        """Provide download header for location to download source."""
        namespace, analytics_token = terrareg.models.Namespace.extract_analytics_token(namespace)

        # If a version has been provided, get the exact version
        if version:
            namespace_obj, module_obj, module_provider_obj, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
            if error:
                return self._get_404_response()
        else:
            # Otherwise, get the module provider, returning an error if it doesn't exist
            namespace_obj, module_obj, module_provider_obj, error = self.get_module_provider_by_names(namespace, name, provider)
            if error:
                return self._get_404_response()
            # Get the latest module version, returning an error if one doesn't exist
            module_version = module_provider_obj.get_latest_version()
            if not module_version:
                return self._get_404_response()

        auth_token = None
        auth_token_match = re.match(r'Bearer (.*)', request.headers.get('Authorization', ''))
        if auth_token_match:
            auth_token = auth_token_match.group(1)

        # Determine if auth token is for internal initialisation of modules
        # during module extraction
        if auth_token == terrareg.config.Config().INTERNAL_EXTRACTION_ANALYTICS_TOKEN:
            pass

        # Determine if auth token is one the auth tokens for ignore anlaytics token check
        elif auth_token in terrareg.config.Config().IGNORE_ANALYTICS_TOKEN_AUTH_KEYS:
            pass

        # otherwise, if module download should be rejected due to
        # non-existent analytics token
        elif not analytics_token and not (terrareg.config.Config().ALLOW_UNIDENTIFIED_DOWNLOADS or terrareg.config.Config().DISABLE_ANALYTICS):
            return make_response(
                ("\nAn {analytics_token_phrase} must be provided.\n"
                 "Please update module source to include {analytics_token_phrase}.\n"
                 "\nFor example:\n  source = \"{host}/{example_analytics_token}__{namespace}/{module_name}/{provider}\"").format(
                    analytics_token_phrase=terrareg.config.Config().ANALYTICS_TOKEN_PHRASE,
                    host=request.host,
                    example_analytics_token=terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN,
                    namespace=namespace_obj.name,
                    module_name=module_obj.name,
                    provider=module_provider_obj.name
                ),
                401
            )
        else:
            # Otherwise, if download is allowed and not internal, record the download
            terrareg.analytics.AnalyticsEngine.record_module_version_download(
                namespace_name=namespace,
                module_name=name,
                provider_name=provider,
                module_version=module_version,
                analytics_token=analytics_token,
                terraform_version=request.headers.get('X-Terraform-Version', None),
                user_agent=request.headers.get('User-Agent', None),
                auth_token=auth_token
            )

        resp = make_response('', 204)
        resp.headers['X-Terraform-Get'] = module_version.get_source_download_url()
        return resp

