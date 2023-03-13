
import hashlib
import hmac
import re

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_version_create
import terrareg.database
import terrareg.config
import terrareg.models
import terrareg.module_extractor
import terrareg.errors


class ApiModuleVersionCreateGitHubHook(ErrorCatchingResource):
    """Provide interface for GitHub hook to detect new and changed releases."""

    def _post(self, namespace, name, provider):
        """Create, update or delete new version based on GitHub release hooks."""
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
            return {'status': 'Error', 'message': 'Module provider is not configured with a repository'}, 400

        github_data = request.json

        if not ('release' in github_data and type(github_data['release']) == dict):
            return {'status': 'Error', 'message': 'Received a non-release hook request'}, 400

        release = github_data['release']

        # Obtain tag name
        tag_ref = release.get('tag_name')
        if not tag_ref:
            return {'status': 'Error', 'message': 'tag_name not present in request'}, 400

        # Attempt to match version against regex
        version = module_provider.get_version_from_tag(tag_ref)

        if not version:
            return {'status': 'Error', 'message': 'Release tag does not match configured version regex'}, 400

        # Create module version
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version=version)

        action = github_data.get('action')
        if not action:
            return {"status": "Error", "message": "No action present in request"}, 400

        if action in ['deleted', 'unpublished']:
            if not terrareg.config.Config().UPLOAD_API_KEYS:
                return {
                    'status': 'Error',
                    'message': 'Version deletion requires API key authentication',
                    'tag': tag_ref
                }, 400
            module_version.delete()

            return {
                'status': 'Success'
            }
        else:
            # Perform import from git
            try:
                with terrareg.module_version_create.module_version_create(module_version):
                    with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as me:
                        me.process_upload()

            except terrareg.errors.TerraregError as exc:
                return {
                    'status': 'Error',
                    'message': 'Tag failed to import',
                    'tag': tag_ref
                }, 500
            else:
                return {
                    'status': 'Success',
                    'message': 'Imported provided tag',
                    'tag': tag_ref
                }
