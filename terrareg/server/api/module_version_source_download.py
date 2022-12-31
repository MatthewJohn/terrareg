
from flask import send_from_directory

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.config


class ApiModuleVersionSourceDownload(ErrorCatchingResource):
    """Return source package of module version"""

    def _get(self, namespace, name, provider, version):
        """Return static file."""
        if not terrareg.config.Config().ALLOW_MODULE_HOSTING:
            return {'message': 'Module hosting is disbaled'}, 500

        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        return send_from_directory(module_version.base_directory, module_version.archive_name_zip)

