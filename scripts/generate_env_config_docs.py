#!python

import re
import sys

sys.path.append('.')

from terrareg.config import Config


valid_config_re = re.compile(r'^[A-Z]')
strip_leading_space_re = re.compile(r'^ +', re.MULTILINE)

for prop in dir(Config):

    # Check if attribute looks like a config variable
    if not valid_config_re.match(prop):
        continue

    default_value = getattr(Config(), prop)

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

    description = getattr(Config, prop).__doc__ or ''
    description = strip_leading_space_re.sub('', description)

    print("""
### {name}

{description}

Default: `{documented_default_value}`

""".format(
    name=prop,
    description=description,
    documented_default_value=documented_default_value))
