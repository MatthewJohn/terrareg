
import pytest


from terrareg.markdown_link_modifier import markdown
from test.unit.terrareg import TerraregUnitTest

class TestMarkdownLinkModifier(TerraregUnitTest):

    @staticmethod
    def _convert_markdown(markdown_text):
        """Return instance of markdown with plugin"""
        return markdown(
            markdown_text,
            file_name="test_file_name",
            extensions=[
                'fenced_code',
                'tables',
                'mdx_truly_sane_lists',
                'terrareg.markdown_link_modifier']
        )

    def test_multiline_table(self):
        """Verify expected behavoir when providing a multi-line table from terraform-docs"""
        assert self._convert_markdown(
            """
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_name"></a> [name](#input\_name) | Enter your name<br>This Should be your full name | `string` | `"My Name"` | no |
| <a name="input_test"></a> [test](#input\_test) | This should contain a large description of a variable.<br>For more informatino glahbab<br><br>adgadgadglkadnladnglakdnglakdgg<br>adgadmad ad gadg adg ad adg adg<br><br>adg<br>a | `string` | n/a | yes |
""".strip()
        ).strip() == """
<h2 id="terrareg-anchor-test_file_name-inputs">Inputs</h2>
<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
<th>Type</th>
<th>Default</th>
<th style="text-align: center;">Required</th>
</tr>
</thead>
<tbody>
<tr>
<td><a name="terrareg-anchor-test_file_name-input_name"></a> <a href="#terrareg-anchor-test_file_name-input95name">name</a></td>
<td>Enter your name<br>This Should be your full name</td>
<td><code>string</code></td>
<td><code>"My Name"</code></td>
<td style="text-align: center;">no</td>
</tr>
<tr>
<td><a name="terrareg-anchor-test_file_name-input_test"></a> <a href="#terrareg-anchor-test_file_name-input95test">test</a></td>
<td>This should contain a large description of a variable.<br>For more informatino glahbab<br><br>adgadgadglkadnladnglakdnglakdgg<br>adgadmad ad gadg adg ad adg adg<br><br>adg<br>a</td>
<td><code>string</code></td>
<td>n/a</td>
<td style="text-align: center;">yes</td>
</tr>
</tbody>
</table>
""".strip()

    def test_placeholders_in_markdown(self):
        """Test use of angular brackets as placeholders in markdown"""
        assert self._convert_markdown("""
Convert `<Hi>` or `<Your name>` where <your-name> is replaced with you name.
        """.strip()) == """
<p>Convert <code>&lt;Hi&gt;</code> or <code>&lt;Your name&gt;</code> where <your-name> is replaced with you name.</p>
""".strip()
