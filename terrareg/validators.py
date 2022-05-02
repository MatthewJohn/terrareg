
from terrareg.errors import RepositoryUrlParseError


class GitUrlValidator:

    def __init__(self, template):
        """Store member variables."""
        self._template = template
        print(template)

    def validate(self,
                 requires_namespace_placeholder=False,
                 requires_module_placeholder=False,
                 requires_tag_placeholder=False,
                 requires_path_placeholder=False):

        really_random_string = 'D3f1N1t3LyW0nt3x15t!'
        if requires_namespace_placeholder:
            if '{namespace}' not in self._template:
                raise RepositoryUrlParseError('Namespace placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace=really_random_string,
                    module='',
                    provider='',
                    path='',
                    tag=''):
                raise RepositoryUrlParseError('Template does not contain valid namespace placeholder')

        if requires_module_placeholder:
            if '{module}' not in self._template:
                raise RepositoryUrlParseError('Module placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace='',
                    module=really_random_string,
                    provider='',
                    path='',
                    tag=''):
                raise RepositoryUrlParseError('Template does not contain valid module placeholder')

        if requires_tag_placeholder:
            if '{tag}' not in self._template:
                raise RepositoryUrlParseError('tag placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace='',
                    module=really_random_string,
                    provider='',
                    path='',
                    tag=really_random_string):
                raise RepositoryUrlParseError('Template does not contain valid tag placeholder')

        if requires_path_placeholder:
            if '{path}' not in self._template:
                raise RepositoryUrlParseError('Path placeholder not present in URL')
            if really_random_string not in self._template.format(
                    namespace='',
                    module='',
                    provider='',
                    path=really_random_string,
                    tag=''):
                raise RepositoryUrlParseError('Template does not contain valid path placeholder')

    def get_value(self, namespace, module, provider, tag, path):
        """Return value with placeholders replaced."""
        return self._format.format(
            namespace=namespace,
            module=module,
            provider=provider,
            tag=tag,
            path=path
        )
