
import unittest.mock
import pytest

from terrareg.models import Example, Module, Namespace, ModuleProvider, ModuleVersion
from test.integration.terrareg.models.test_base_submodule import CommonBaseSubmodule


class TestExample(CommonBaseSubmodule):

    SUBMODULE_CLASS = Example

    @pytest.mark.parametrize('readme_content,example_analytics_token,expected_output', [
        # Test README with basic formatting
        (
            """
# Test terraform module

This is a terraform module to create a README example.

It performs the following:

 * Creates a README
 * Tests the README
 * Passes tests
""",
            "unittest-analytics-token",
            """
<h1>Test terraform module</h1>
<p>This is a terraform module to create a README example.</p>
<p>It performs the following:</p>
<ul>
<li>Creates a README</li>
<li>Tests the README</li>
<li>Passes tests</li>
</ul>
"""
        ),
        # Test README with external module call
        (
            """
# Test external module

```
module "test-usage" {
  source  = "an-external-module/test"
  version = "1.0.0"

  some_variable = true
  another       = "value"
}
```
""",
            "unittest-analytics-token",
            """
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;an-external-module/test&quot;
  version = &quot;1.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),
        # Test README with call to current example
        (
            """
# Test external module

```
module "test-usage" {
  source  = "./"

  some_variable = true
  another       = "value"
}
```
            """,
            "unittest-analytics-token",
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//examples/testreadmeexample&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test call without analytics token
        (
            """
# Test external module

```
module "test-usage" {
  source  = "./"

  some_variable = true
  another       = "value"
}
```
            """,
            None,
"""
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/moduledetails/readme-tests/provider//examples/testreadmeexample&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test README with call outside of module root
        (
            """
# Test external module

```
module "test-usage" {
  source  = "../../"

  some_variable = true
  another       = "value"
}
```
""",
            "unittest-analytics-token",
            """
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test with call to submodule
        (
            """
# Test external module

```
module "test-usage" {
  source  = "../../modules/testsubmodule"

  some_variable = true
  another       = "value"
}
```
""",
            "unittest-analytics-token",
            """
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//modules/testsubmodule&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test with call to submodule within example
        (
            """
# Test external module

```
module "test-usage" {
  source  = "./testexamplesubmodule"

  some_variable = true
  another       = "value"
}
```
""",
            "unittest-analytics-token",
            """
<h1>Test external module</h1>
<pre><code>module &quot;test-usage&quot; {
  source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//examples/testreadmeexample/testexamplesubmodule&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
</code></pre>
"""
        ),

        # Test README with multiple modules
        (
            """
# Test external module

```
module "test-usage1" {
  source = "./"

  some_variable = true
  another       = "value"
}
module "test-usage2" {
  source = "../../modules/testsubmodule"

  some_variable = true
  another       = "value"
}
module "test-external-call" {
  source  = "external-module"
  version = "1.0.3"
}
```
""",
            "unittest-analytics-token",
            """
<h1>Test external module</h1>
<pre><code>module &quot;test-usage1&quot; {
  source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//examples/testreadmeexample&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
module &quot;test-usage2&quot; {
  source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//modules/testsubmodule&quot;
  version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;

  some_variable = true
  another       = &quot;value&quot;
}
module &quot;test-external-call&quot; {
  source  = &quot;external-module&quot;
  version = &quot;1.0.3&quot;
}
</code></pre>
"""
        ),

        # Test module call with different indentation
        (
            """
# Test external module

```
module "test-usage1" {
  source        = "./"
  some_variable = true
  another       = "value"
}
module "test-usage2" {
    source = "../anotherexample"
}
module "test-usage3" {
          source =         "././.././../modules/testsubmodule"
}
```
""",
            "unittest-analytics-token",
            """
<h1>Test external module</h1>
<pre><code>module &quot;test-usage1&quot; {
  source        = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//examples/testreadmeexample&quot;
  version       = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;
  some_variable = true
  another       = &quot;value&quot;
}
module &quot;test-usage2&quot; {
    source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//examples/anotherexample&quot;
    version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;
}
module &quot;test-usage3&quot; {
          source  = &quot;example.com/unittest-analytics-token__moduledetails/readme-tests/provider//modules/testsubmodule&quot;
          version = &quot;&gt;= 1.0.0, &lt; 2.0.0&quot;
}
</code></pre>
"""
        ),
    ])
    def test_get_readme_html(self, readme_content, example_analytics_token, expected_output):
        """Test get_readme_html method of example, ensuring it replaces example source and converts from markdown to HTML."""

        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_EXAMPLE_VERSION_TEMPLATE', '>= {major}.{minor}.{patch}, < {major_plus_one}.0.0'), \
                unittest.mock.patch('terrareg.config.Config.EXAMPLE_ANALYTICS_TOKEN', example_analytics_token):
            module_version = ModuleVersion(ModuleProvider(Module(Namespace('moduledetails'), 'readme-tests'), 'provider'), '1.0.0')
            example = Example(module_version=module_version, module_path='examples/testreadmeexample')
            # Set README in module version
            example.module_details.update_attributes(readme_content=readme_content)

            assert example.get_readme_html(server_hostname='example.com').strip() == expected_output.strip()
