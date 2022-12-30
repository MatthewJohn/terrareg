
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.config
import terrareg.models
import terrareg.database
import terrareg.errors
import terrareg.module_extractor


class ApiModuleVersionUpload(ErrorCatchingResource):

    ALLOWED_EXTENSIONS = ['zip']

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_upload_module_version', request_kwarg_map={'namespace': 'namespace'})]

    def allowed_file(self, filename):
        """Check if file has allowed file-extension"""
        return '.' in filename and \
               filename.rsplit('.', 1)[1].lower() in self.ALLOWED_EXTENSIONS

    def _post(self, namespace, name, provider, version):
        """Handle module version upload."""

        # If module hosting is disabled,
        # refuse the upload of new modules
        if not terrareg.config.Config().ALLOW_MODULE_HOSTING:
            return {'message': 'Module upload is disabled.'}, 400

        with terrareg.database.Database.start_transaction():

            # Get module provider and, optionally create, if it doesn't exist
            _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider, create=True)
            if error:
                return error

            module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version=version)

            if len(request.files) != 1:
                raise terrareg.errors.UploadError('One file can be uploaded')

            file = request.files[[f for f in request.files.keys()][0]]

            # If the user does not select a file, the browser submits an
            # empty file without a filename.
            if file.filename == '':
                raise terrareg.errors.UploadError('No selected file')

            if not file or not self.allowed_file(file.filename):
                raise terrareg.errors.UploadError('Error occurred - unknown file extension')

            module_version.prepare_module()
            with terrareg.module_extractor.ApiUploadModuleExtractor(upload_file=file, module_version=module_version) as me:
                me.process_upload()

            return {
                'status': 'Success'
            }
