
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiModuleVersions(ErrorCatchingResource):

    def _get(self, namespace, name, provider):
        """Return list of version."""

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)
        namespace, module, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return self._get_404_response()

        return {
            "modules": [
                {
                    "source": "{namespace}/{module}/{provider}".format(
                        namespace=namespace.name,
                        module=module.name,
                        provider=module_provider.name
                    ),
                    "versions": [
                        {
                            "version": v.version,
                            "root": {
                                # @TODO: Add providers/depdencies
                                "providers": [],
                                "dependencies": []
                            },
                            # @TODO: Add submodule information
                            "submodules": []
                        }
                        for v in module_provider.get_versions()
                    ]
                }
            ]
        }
