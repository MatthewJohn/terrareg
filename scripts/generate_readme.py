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
    newline = "\n"
    api_docs += f"""


## {route_class.__name__}

{newline.join([f'`{url}`' for url in urls])}

{route_class.__doc__ or ""}

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

        api_docs += f'\n### {method}\n\n{getattr(route_class, internal_method).__doc__ or ""}'


server._app.route = unittest.mock.MagicMock()
server._api.add_resource = unittest.mock.MagicMock(side_effect=mock_route)
server._register_routes()

with open('docs/API.md', 'w') as readme:
    readme.write(api_docs)

# initial attempt
# api_docs_config = ""
# for rule in server._app.url_map.iter_rules():
#     if rule.endpoint == 'static':
#         continue
#     print(rule)
#     print(rule.methods)
#     print(rule.endpoint)
#     print(getattr(server._app.view_functions[rule.endpoint], "GET", None))
#     api_docs_config += f"""
# ## {rule.methods}
# """

