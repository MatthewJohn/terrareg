
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

| Table       | Heading |
| ----------- | ------- |
| Row2b       | Row1    |
| Row2a       | Row2b   |
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
<table>
<thead>
<tr>
<th>Table</th>
<th>Heading</th>
</tr>
</thead>
<tbody>
<tr>
<td>Row2b</td>
<td>Row1</td>
</tr>
<tr>
<td>Row2a</td>
<td>Row2b</td>
</tr>
</tbody>
</table>

""".strip()
        ),

        (
            'randomcharacters.md',
            """
# Heading With 123 !"£$%^&*()_+}{][~@:#';?></.,

# Pre-existing_hypens-and_underscores

Should only show alphanumeric in ID
""",
            """
<h1 id="terrareg-anchor-randomcharactersmd-heading-with-123-_">Heading With 123 !"£$%^&amp;*()_+}{][~@:#';?&gt;&lt;/.,</h1>
<h1 id="terrareg-anchor-randomcharactersmd-pre-existing_hypens-and_underscores">Pre-existing_hypens-and_underscores</h1>
<p>Should only show alphanumeric in ID</p>
""".strip()
        ),

        (
            'linkreplacement.md',
            """
[Just anchor](#some-heading)
[With filename](linkreplacement.md#some-heading)
[With leading slash](./linkreplacement.md#some-heading)

[Another MD file](anotherfile.md#some-heading)
[Another MD file with leading slash](./anotherfile.md#some-heading)

[A different file](../source/tags/blah)

[External link](https://example.com)
""",
            """
<p><a href="#terrareg-anchor-linkreplacementmd-some-heading">Just anchor</a>
<a href="#terrareg-anchor-linkreplacementmd-some-heading">With filename</a>
<a href="#terrareg-anchor-linkreplacementmd-some-heading">With leading slash</a></p>
<p><a href="anotherfile.md#some-heading">Another MD file</a>
<a href="./anotherfile.md#some-heading">Another MD file with leading slash</a></p>
<p><a href="../source/tags/blah">A different file</a></p>
<p><a href="https://example.com">External link</a></p>
""".strip()
        ),

        (
            'embeddedhtml.md',
            """
<a href="https://www.github.com">Github</a>
<a href="#someanchor">Just Anchor</a>
<a href="embeddedhtml.md#someanchor">MD File</a>
<a href="./embeddedhtml.md#someanchor">MD File trailing leading slash</a>
<a href="readme.md#anotheranchor">Different MD file</a>
<a href="./readme.md#anotheranchor">Different MD file leading slash</a>

<a id="test-id">With ID</a>
<a name="test-name">With Name</a>
""",
            """
<p><a href="https://www.github.com">Github</a>
<a href="#terrareg-anchor-embeddedhtmlmd-someanchor">Just Anchor</a>
<a href="#terrareg-anchor-embeddedhtmlmd-someanchor">MD File</a>
<a href="#terrareg-anchor-embeddedhtmlmd-someanchor">MD File trailing leading slash</a>
<a href="readme.md#anotheranchor">Different MD file</a>
<a href="./readme.md#anotheranchor">Different MD file leading slash</a></p>
<p><a id="terrareg-anchor-embeddedhtmlmd-test-id">With ID</a>
<a name="terrareg-anchor-embeddedhtmlmd-test-name">With Name</a></p>
""".strip()
        ),

        # Test image source replacement for relative image paths
        (
            "testimages.md",
            """
![Actual image](relativeimage.png)

![Relative Path](./relativeimage.png)

![Absolute Path](/img/absolutepath.png)

![external http](http://example.com/myimage.png)

![external https](https://example.com/myimage.png)
""",
            """
<p><img></p>
<p><img></p>
<p><img></p>
<p><img src="http://example.com/myimage.png"></p>
<p><img src="https://example.com/myimage.png"></p>
""".strip()
        ),

        # With HTML injection
        (
            'injection.md',
            '<script>console.log("Hi");</script>',
            '&lt;script&gt;console.log("Hi");&lt;/script&gt;'
        )
    ])
    def test_module_version_file_content(self, file_name, file_content, expected_output):
        """Test source replacement in example file content."""

        module_version = ModuleVersion(ModuleProvider(Module(Namespace('moduledetails'), 'readme-tests'), 'provider'), '1.0.0')
        module_version_file = ModuleVersionFile.create(module_version, file_name)
        module_version_file.update_attributes(content=file_content)
        assert module_version_file.get_content() == expected_output
