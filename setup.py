#!/usr/bin/env python

from distutils.core import setup
from pip.req import parse_requirements

install_reqs = parse_requirements('requirements.txt', session='hack')

setup(name='terrareg',
      version='0.1.0',
      description='Terraform module regsitry with analytics',
      author='Matt Comben',
      author_email='matthew@dockstudios.co.uk',
      url='https://gitlab.dockstudios.co.uk/pub/terrareg',
      packages=['terrareg'],
      install_requires=reqs
)
