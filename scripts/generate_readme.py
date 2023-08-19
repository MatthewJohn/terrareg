#!python

from enum import EnumMeta
import re
import sys
import os
import unittest.mock

import jinja2

sys.path.append('.')

from terrareg.config import Config
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.server import Server


valid_config_re = re.compile(r'^[A-Z]')
strip_leading_space_re = re.compile(r'^ +', re.MULTILINE)

config_contents = ""

for prop in dir(Config):

    # Check if attribute looks like a config variable
    if not valid_config_re.match(prop):
        continue

    default_value = getattr(Config(), prop)

    # Convert enum types to the value of the default enum
    if type(default_value.__class__) is EnumMeta:
        default_value = default_value.value

    # If the config becomes a list,
    # default to empty string, as it will
    # be a comma-separated list
    if default_value == [] or default_value is None:
        documented_default_value = ''

    elif default_value is True:
        documented_default_value = 'True'
    elif default_value is False:
        documented_default_value = 'False'
    else:
        documented_default_value = default_value

    if type(documented_default_value) is str and os.getcwd() in documented_default_value:
        documented_default_value = documented_default_value.replace(os.getcwd(), '.')

    description = getattr(Config, prop).__doc__ or ''
    description = strip_leading_space_re.sub('', description)

    config_contents += """
### {name}

{description}

Default: `{documented_default_value}`

""".format(
    name=prop,
    description=description,
    documented_default_value=documented_default_value)

with open('docs/CONFIG.md.in', 'r') as readme_in:
    readme_template = ''.join(readme_in.readlines())

template = jinja2.Template(readme_template)
readme_out = template.render(CONFIG_CONTENTS=config_contents)

with open('docs/CONFIG.md', 'w') as readme:
    readme.write(readme_out)


server = Server()

api_docs = """
# API Docs

"""

def mock_route(route_class, *urls):
    global api_docs
    newline = "\n\n"
    class_docs = route_class.__doc__ or ""
    class_docs = "\n".join([l.strip() for l in class_docs.split("\n")])

    api_docs += f"""


## {route_class.__name__}

{newline.join([f'`{url}`' for url in urls])}

{class_docs}

"""

    for method in ['GET', 'POST', 'DELETE']:
        internal_method = method.lower()

        # Check if method exists
        if not hasattr(route_class, internal_method):
            continue

        # If class is a subclass of ErrorCatchingResource,
        # skip method if it does not override the ErrorCatchingResource base method
        if route_class in ErrorCatchingResource.__subclasses__():
            internal_method = f'_{method.lower()}'
            if getattr(route_class, internal_method) == getattr(ErrorCatchingResource, internal_method):
                continue

        method_docs = getattr(route_class, internal_method).__doc__ or ""
        method_docs = "\n".join([l.strip() for l in method_docs.split("\n")])

        api_docs += f'\n### {method}\n\n{method_docs}'

        # Attempt to get arg parser
        if route_class in ErrorCatchingResource.__subclasses__():
            arg_parser_method = f"_{method.lower()}_arg_parser"
            if getattr(route_class, arg_parser_method) != getattr(ErrorCatchingResource, arg_parser_method):
                api_docs += """
#### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
"""
                arg_parser = getattr(route_class(), arg_parser_method)()
                for arg in arg_parser.args:
                    api_docs += f'| {arg.name} | {arg.location} | {arg.type.__name__ if arg.type else ""} | {arg.required} | `{arg.default}` | {arg.help if arg.help else ""} |\n'


server._app.route = unittest.mock.MagicMock()
server._api.add_resource = unittest.mock.MagicMock(side_effect=mock_route)
server._register_routes()

with open('docs/API.md', 'w') as readme:
    readme.write(api_docs)
