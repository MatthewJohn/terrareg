#!/usr/bin/env python

from setuptools import setup

with open('requirements.txt') as f:
    required = f.read().splitlines()

setup(name='terrareg',
      version='0.1.0',
      description='Terraform module regsitry with analytics',
      author='Matt Comben',
      author_email='matthew@dockstudios.co.uk',
      url='https://gitlab.dockstudios.co.uk/pub/terrareg',
      packages=['terrareg'],
      install_requires=required
)
