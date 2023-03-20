
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

### Heading 3

#### Heading 4

##### Heading 5

###### Heading 6

 * Some list
 * Another list item
""",
            """
<h1 id="terrareg-anchor-testmd-heading-1">Heading 1</h1>
<h2 id="terrareg-anchor-testmd-heading-2">Heading 2</h2>
<h3 id="terrareg-anchor-testmd-heading-3">Heading 3</h3>
<h4 id="terrareg-anchor-testmd-heading-4">Heading 4</h4>
<h5 id="terrareg-anchor-testmd-heading-5">Heading 5</h5>
<h6 id="terrareg-anchor-testmd-heading-6">Heading 6</h6>
<ul>
<li>Some list</li>
<li>Another list item</li>
</ul> 
""".strip()
        ),

        (
            'randomcharacters.md',
            """
# heading with !"£$%^&*()_+}{][~@:#';?></.,]}

Should only show alphanumeric in ID
""",
            """
<h1 id="terrareg-anchor-randomcharactersmd-heading-with-">heading with !"£$%^&amp;*()_+}{][~@:#';?&gt;&lt;/.,]}</h1>
<p>Should only show alphanumeric in ID</p>
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
