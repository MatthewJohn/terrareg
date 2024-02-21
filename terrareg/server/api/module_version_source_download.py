

import os
from flask import send_file, request

from terrareg.errors import InvalidPresignedUrlKeyError
from terrareg.presigned_url import TerraformSourcePresignedUrl
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.config
import terrareg.file_storage


class ApiModuleVersionSourceDownload(ErrorCatchingResource):
    """Return source package of module version"""

    def _get(self, namespace, name, provider, version):
        """Return static file."""
        config = terrareg.config.Config()
        if not config.ALLOW_MODULE_HOSTING:
            return {'message': 'Module hosting is disbaled'}, 500

        # If authentication is required, check pre-signed URL
        if not config.ALLOW_UNAUTHENTICATED_ACCESS:
            try:
                TerraformSourcePresignedUrl.validate_presigned_key(url=request.path, payload=request.args.get("presign"))
            except InvalidPresignedUrlKeyError as exc:
                return {'message': str(exc)}, 403

        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        file_storage = terrareg.file_storage.FileStorageFactory().get_file_storage()
        return send_file(
            file_storage.read_file(os.path.join(module_version.base_directory, module_version.archive_name_zip), bytes_mode=True),
            download_name=module_version.archive_name_zip
        )
