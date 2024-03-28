
import urllib.parse

from terrareg.errors import RepositoryUrlParseError, RepositoryUrlContainsInvalidTemplateError


class GitUrlValidator:

    def __init__(self, template):
        """Store member variables."""
        self._template = template

    def validate(self,
                 requires_namespace_placeholder=False,
                 requires_module_placeholder=False,
                 requires_tag_placeholder=False,
                 requires_path_placeholder=False):

        # Ensure exceptions are not thrown when formatting string
        try:
            self._template.format(
                namespace='',
                module='',
                provider='',
                path='',
                tag='',
                tag_uri_encoded=''
            )
        except KeyError as exc:
            raise RepositoryUrlContainsInvalidTemplateError(
                f"Template contains unknown placeholder: {', '.join(exc.args)}. "
                "Valid placeholders are contain: {namespace}, {module}, {provider}, {path}, {tag} and {tag_uri_encoded}"
            )

        really_random_string = 'D3f1N1t3LyW0nt3x15t!'
        if requires_namespace_placeholder:
            if '{namespace}' not in self._template:
                raise RepositoryUrlParseError('Namespace placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace=really_random_string,
                    module='',
                    provider='',
                    path='',
                    tag='',
                    tag_uri_encoded=''):
                raise RepositoryUrlParseError('Template does not contain valid namespace placeholder')

        if requires_module_placeholder:
            if '{module}' not in self._template:
                raise RepositoryUrlParseError('Module placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace='',
                    module=really_random_string,
                    provider='',
                    path='',
                    tag='',
                    tag_uri_encoded=''):
                raise RepositoryUrlParseError('Template does not contain valid module placeholder')

        if requires_tag_placeholder:
            if '{tag}' not in self._template and '{tag_uri_encoded}' not in self._template:
                raise RepositoryUrlParseError('tag or tag_uri_encoded placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace='',
                    module='',
                    provider='',
                    path='',
                    tag=really_random_string,
                    tag_uri_encoded=really_random_string):
                raise RepositoryUrlParseError('Template does not contain valid tag/tag_uri_encoded placeholder')

        if requires_path_placeholder:
            if '{path}' not in self._template:
                raise RepositoryUrlParseError('Path placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace='',
                    module='',
                    provider='',
                    path=really_random_string,
                    tag='',
                    tag_uri_encoded=''):
                raise RepositoryUrlParseError('Template does not contain valid path placeholder')

    def get_value(self, namespace, module, provider, tag, path):
        """Return value with placeholders replaced."""
        return self._template.format(
            namespace=namespace,
            module=module,
            provider=provider,
            tag=tag,
            tag_uri_encoded=urllib.parse.quote(tag, safe=''),
            path=path
        )
