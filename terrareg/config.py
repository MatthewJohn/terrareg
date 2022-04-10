
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
