
import hashlib
import hmac
import re

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.database
import terrareg.config
import terrareg.models
import terrareg.module_extractor
import terrareg.errors


class ApiModuleVersionCreateBitBucketHook(ErrorCatchingResource):
    """Provide interface for Bitbucket hook to detect pushes of new tags."""

    def _post(self, namespace, name, provider):
        """Create new version based on Bitbucket hooks."""
        with terrareg.database.Database.start_transaction() as transaction_context:
            _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
            if error:
                return error

            # Validate signature
            if terrareg.config.Config().UPLOAD_API_KEYS:
                # Get signature from request
                request_signature = request.headers.get('X-Hub-Signature', '')
                # Remove 'sha256=' from beginning of header
                request_signature = re.sub(r'^sha256=', '', request_signature)
                # Iterate through each of the keys and test
                for test_key in terrareg.config.Config().UPLOAD_API_KEYS:
                    # Generate
                    valid_signature = hmac.new(bytes(test_key, 'utf8'), b'', hashlib.sha256)
                    valid_signature.update(request.data)
                    # If the signatures match, break from loop
                    if hmac.compare_digest(valid_signature.hexdigest(), request_signature):
                        break
                # If a valid signature wasn't found with one of the configured keys,
                # return 401
                else:
                    return self._get_401_response()

            if not module_provider.get_git_clone_url():
                return {'message': 'Module provider is not configured with a repository'}, 400

            bitbucket_data = request.json

            if not ('changes' in bitbucket_data and type(bitbucket_data['changes']) == list):
                return {'message': 'List of changes not found in payload'}, 400

            imported_versions = {}
            error = False

            for change in bitbucket_data['changes']:

                # Check that change is a dict
                if not type(change) is dict:
                    continue

                # Check if change type is tag
                if not ('ref' in change and
                        type(change['ref']) is dict and
                        'type' in change['ref'] and
                        type(change['ref']['type']) == str and
                        change['ref']['type'] == 'TAG'):
                    continue

                # Check type of change is an ADD or UPDATE
                if not ('type' in change and
                        type(change['type']) is str and
                        change['type'] in ['ADD', 'UPDATE']):
                    continue

                # Obtain tag name
                tag_ref = change['ref']['id'] if 'id' in change['ref'] else None

                # Attempt to match version against regex
                version = module_provider.get_version_from_tag_ref(tag_ref)

                if not version:
                    continue

                # Create module version
                module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version=version)

                # Perform import from git
                savepoint = transaction_context.connection.begin_nested()
                try:
                    with module_version.module_create_extraction_wrapper():
                        with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as me:
                            me.process_upload()

                except terrareg.errors.TerraregError as exc:
                    # Roll back the transaction for this module version
                    savepoint.rollback()

                    imported_versions[version] = {
                        'status': 'Failed',
                        'message': str(exc)
                    }
                    error = True
                else:
                    # Commit the transaction for this version
                    savepoint.commit()
                    imported_versions[version] = {
                        'status': 'Success'
                    }

            if error:
                return {
                    'status': 'Error',
                    'message': 'One or more tags failed to import',
                    'tags': imported_versions
                }, 500
            return {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': imported_versions
            }
