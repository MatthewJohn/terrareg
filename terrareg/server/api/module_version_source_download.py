

import os
import flask

from terrareg.errors import InvalidPresignedUrlKeyError
import terrareg.presigned_url
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.config
import terrareg.file_storage


class ApiModuleVersionSourceDownload(ErrorCatchingResource):
    """Return source package of module version"""

    def _get(self, namespace, name, provider, version, presign=None):
        """Return static file."""
        config = terrareg.config.Config()
        if config.ALLOW_MODULE_HOSTING is terrareg.config.ModuleHostingMode.DISALLOW:
            return {'message': 'Module hosting is disbaled'}, 500

        # If authentication is required, check pre-signed URL
        if not config.ALLOW_UNAUTHENTICATED_ACCESS or config.REQUIRE_PRESIGNED_URL_AUTHENTICATION:
            presign = flask.request.args.get("presign", presign)

            path = flask.request.path
            path_parts = path.split('/')
            # Remove last section of path (i.e. the file name)
            del path_parts[-1]

            # If path ends with the pre-sign key, remove it
            if presign and path_parts[-1] == presign:
                del path_parts[-1]
            path = '/'.join(path_parts)

            try:
                terrareg.presigned_url.TerraformSourcePresignedUrl.validate_presigned_key(url=path, payload=presign)
            except InvalidPresignedUrlKeyError as exc:
                return {'message': str(exc)}, 403

        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        file_storage = terrareg.file_storage.FileStorageFactory().get_file_storage()
        return flask.send_file(
            file_storage.read_file(os.path.join(module_version.base_directory, module_version.archive_name_zip), bytes_mode=True),
            download_name=module_version.archive_name_zip,
            mimetype='application/zip'
        )
