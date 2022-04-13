
import os

DATA_DIRECTORY = os.path.join(os.environ.get('DATA_DIRECTORY', os.getcwd()), 'data')

"""
Whether modules can be downloaded with terraform
without specifying an identification string in
the namespace
"""
ALLOW_UNIDENTIFIED_DOWNLOADS = False

"""Whether flask and sqlalchemy is setup in debug mode."""
DEBUG = bool(os.environ.get('DEBUG', False))

"""Name of analytics token to provide in responses (e.g. application name, team name etc.)"""
ANALYTICS_TOKEN_PHRASE = os.environ.get('ANALYTICS_TOKEN_PHRASE', 'analytics token')

"""Example analytics token to provide in responses (e.g. my-tf-application, my-slack-channel etc.)"""
EXAMPLE_ANALYTICS_TOKEN = os.environ.get('EXAMPLE_ANALYTICS_TOKEN', 'my-tf-application')

"""Comma-separated list of trusted namespaces."""
TRUSTED_NAMESPACES = os.environ.get('TRUSTED_NAMESPACES', '').split(',')

"""
Whether uploaded modules, that provide an external URL for the artifact,
should be removed after analysis.
If enabled, module versions with externally hosted artifacts cannot be re-analysed after upload. 
"""
DELETE_EXTERNALLY_HOSTED_ARTIFACTS = os.environ.get('DELETE_EXTERNALLY_HOSTED_ARTIFACTS', 'False') == 'True'
