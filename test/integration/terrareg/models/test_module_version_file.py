
import os
import unittest

import pytest

from terrareg.models import Example, ExampleFile, Module, ModuleVersion, ModuleVersionFile, Namespace, ModuleProvider
from test.integration.terrareg import TerraregIntegrationTest
from test.integration.terrareg.module_extractor import UploadTestModule

class TestModuleVersionFile(TerraregIntegrationTest):

    @pytest.mark.parametrize('file_name,file_content,expected_output', [
        # Test empty file
        ('test_file', '', '<pre></pre>'),
        ('test_file.md', '', ''),

        # File without extension
        (
            'license',
            'This is a test license file\nLicense here',
            '<pre>This is a test license file\nLicense here</pre>'
        ),

        # Markdown file
        (
            'test.md',
            """
# Heading 1

## Heading 2

 * Some list
 * Another list item
""",
            """
<h1>Heading 1</h1>
<h2>Heading 2</h2>
<ul>
<li>Some list</li>
<li>Another list item</li>
</ul> 
""".strip()
        ),

        # With HTML injection
        (
            'injection.md',
            '<script>console.log("Hi");</script>',
            '<p>&lt;script&gt;console.log("Hi");&lt;/script&gt;</p>'
        )
    ])
    def test_module_version_file_content(self, file_name, file_content, expected_output):
        """Test source replacement in example file content."""

        module_version = ModuleVersion(ModuleProvider(Module(Namespace('moduledetails'), 'readme-tests'), 'provider'), '1.0.0')
        module_version_file = ModuleVersionFile.create(module_version, file_name)
        module_version_file.update_attributes(content=file_content)
        assert module_version_file.get_content() == expected_output
