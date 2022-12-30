
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Namespace


class ApiTerraregInitialSetupData(ErrorCatchingResource):
    """Interface to provide data to the initial setup page."""

    def _get(self):
        """Return information for steps for setting up Terrareg."""
        # Get first namespace, if present
        namespace = None
        module = None
        module_provider = None
        namespaces = Namespace.get_all(only_published=False)
        version = None
        integrations = {}
        if namespaces:
            namespace = namespaces[0]
        if namespace:
            modules = namespace.get_all_modules()
            if modules:
                module = modules[0]
        if module:
            providers = module.get_providers()
            if providers:
                module_provider = providers[0]
                integrations = module_provider.get_integrations()

        if module_provider:
            versions = module_provider.get_versions(include_beta=True, include_unpublished=True)
            if versions:
                version = versions[0]

        return {
            "namespace_created": bool(namespaces),
            "module_created": bool(module_provider),
            "version_indexed": bool(version),
            "version_published": bool(version.published) if version else False,
            "module_configured_with_git": bool(module_provider.get_git_clone_url()) if module_provider else False,
            "module_view_url": module_provider.get_view_url() if module_provider else None,
            "module_upload_endpoint": integrations['upload']['url'] if 'upload' in integrations else None,
            "module_publish_endpoint": integrations['publish']['url'] if 'publish' in integrations else None
        }