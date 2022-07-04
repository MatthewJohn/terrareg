
from time import sleep
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.select import Select

from test.selenium import SeleniumTest
from terrareg.models import GitProvider, ModuleVersion, Namespace, Module, ModuleProvider

class TestModuleProvider(SeleniumTest):
    """Test module provider page."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls._api_version_create_mock = mock.Mock(return_value={'status': 'Success'})
        cls._api_version_publish_mock = mock.Mock(return_value={'status': 'Success'})
        cls._config_publish_api_keys_mock = mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', [])
        cls._config_allow_custom_repo_urls_module_provider = mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True)

        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))
        cls.register_patch(mock.patch('terrareg.config.Config.SECRET_KEY', '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'))
        cls.register_patch(mock.patch('terrareg.server.ApiModuleVersionCreate._post', cls._api_version_create_mock))
        cls.register_patch(mock.patch('terrareg.server.ApiTerraregModuleVersionPublish._post', cls._api_version_publish_mock))
        cls.register_patch(cls._config_publish_api_keys_mock)
        cls.register_patch(cls._config_allow_custom_repo_urls_module_provider)

        super(TestModuleProvider, cls).setup_class()

    def test_module_without_versions(self):
        """Test page functionality on a module without published versions."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/noversion/testprovider'))

        # Ensure integrations tab link is display and tab is displayed
        self.wait_for_element(By.ID, 'module-tab-link-integrations')
        integration_tab = self.wait_for_element(By.ID, 'module-tab-integrations')

        # Ensure all other tabs aren't shown
        for tab_name in ['readme', 'example-files', 'inputs', 'outputs', 'providers', 'resources', 'analytics', 'usage-builder', 'settings']:
            # Ensure tab link isn't displayed
            assert self.selenium_instance.find_element(By.ID, f'module-tab-link-{tab_name}').is_displayed() == False

            # Ensure tab content isn't displayed
            assert self.selenium_instance.find_element(By.ID, f'module-tab-{tab_name}').is_displayed() == False

        # Login
        self.perform_admin_authentication(password='unittest-password')

        # Ensure settings tab link is displayed
        self.selenium_instance.get(self.get_url('/modules/moduledetails/noversion/testprovider'))

        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, f'module-tab-link-settings').is_displayed(), True)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, f'module-tab-settings').is_displayed(), False)

    @pytest.mark.parametrize('url,expected_readme_content', [
        # Root module
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0', 'This is an exaple README!'),
        # Module example
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example', 'Example 1 README'),
        # Submodule
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1', 'Submodule 1 README')
    ])
    def test_readme_tab(self, url, expected_readme_content):
        """Ensure README is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Ensure README link and tab is displayed by default
        self.wait_for_element(By.ID, 'module-tab-link-readme')
        readme_content = self.wait_for_element(By.ID, 'module-tab-readme')

        # Check contents of REAMDE
        assert readme_content.text == expected_readme_content

        # Navigate to another tab
        self.selenium_instance.find_element(By.ID, 'module-tab-link-inputs').click()

        # Ensure that README content is no longer visible
        assert self.selenium_instance.find_element(By.ID, 'module-tab-readme').is_displayed() == False

        # Click on README tab again
        self.selenium_instance.find_element(By.ID, 'module-tab-link-readme').click()

        # Ensure README content is visible again and content is correct
        readme_content = self.selenium_instance.find_element(By.ID, 'module-tab-readme')
        assert readme_content.is_displayed() == True
        assert readme_content.text == expected_readme_content

    @pytest.mark.parametrize('url,expected_inputs', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['name_of_application', 'Enter the application name', 'string', 'Required'],
                ['string_with_default_value', 'Override the default string', 'string', '"this is the default"'],
                ['example_boolean_input', 'Override the truthful boolean', 'bool', 'true'],
                ['example_list_input', 'Override the stringy list', 'list', '["value 1","value 2"]']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['input_for_example', 'Enter the example name', 'string', 'Required']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['input_for_submodule', 'Enter the submodule name', 'string', 'Required']]
        )
    ])
    def test_inputs_tab(self, url, expected_inputs):
        """Ensure inputs tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for input tab button to be visible
        input_tab_button = self.wait_for_element(By.ID, 'module-tab-link-inputs')

        # Ensure the inputs tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-inputs', ensure_displayed=False).is_displayed() == False

        # Click on inputs tab
        input_tab_button.click()

        # Obtain tab content
        inputs_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-inputs')

        # Ensure tab is displayed
        self.assert_equals(lambda: inputs_tab_content.is_displayed(), True)

        inputs_table = inputs_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = inputs_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_inputs) + 1)

        for row_itx, expected_row in enumerate([['Name', 'Description', 'Type', 'Default value']] + expected_inputs):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 4 columns
            assert len(row_columns) == 4

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    @pytest.mark.parametrize('url,expected_outputs', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['generated_name', 'Name with randomness'],
                ['no_desc_output', '']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['example_output', 'Example name with randomness']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['submodule_output', 'Submodule name with randomness']]
        )
    ])
    def test_outputs_tab(self, url, expected_outputs):
        """Ensure outputs tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for outputs tab button to be visible
        outputs_tab_button = self.wait_for_element(By.ID, 'module-tab-link-outputs')

        # Ensure the outputs tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-outputs', ensure_displayed=False).is_displayed() == False

        # Click on outputs tab
        outputs_tab_button.click()

        # Obtain tab content
        outputs_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-outputs')

        # Ensure tab is displayed
        self.assert_equals(lambda: outputs_tab_content.is_displayed(), True)

        outputs_table = outputs_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = outputs_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_outputs) + 1)

        for row_itx, expected_row in enumerate([['Name', 'Description']] + expected_outputs):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 2 columns
            assert len(row_columns) == 2

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    @pytest.mark.parametrize('url,expected_resources', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['string', 'random_suffix', 'random', 'hashicorp/random', 'managed', 'latest', '']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['string', 'example_random_suffix', 'example_random', 'hashicorp/example_random', 'managed', 'latest', '']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['string', 'submodule_random_suffix', 'submodule_random', 'hashicorp/submodule_random', 'managed', 'latest', '']]
        )
    ])
    def test_resources_tab(self, url, expected_resources):
        """Ensure resources tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for resources tab button to be visible
        resources_tab_button = self.wait_for_element(By.ID, 'module-tab-link-resources')

        # Ensure the resources tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-resources', ensure_displayed=False).is_displayed() == False

        # Click on resources tab
        resources_tab_button.click()

        # Obtain tab content
        resources_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-resources')

        # Ensure tab is displayed
        self.assert_equals(lambda: resources_tab_content.is_displayed(), True)

        resources_table = resources_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = resources_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_resources) + 1)

        for row_itx, expected_row in enumerate(
                [['Type', 'Name', 'Provider', 'Source', 'Mode', 'Version', 'Description']] +
                expected_resources):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 7 columns
            assert len(row_columns) == 7

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    @pytest.mark.parametrize('url,expected_providers', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['random', 'hashicorp', '', '5.2.1'],
                ['unsafe', 'someothercompany', '', '2.0.0']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['example_random', 'hashicorp', '', '']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['submodule_random', 'hashicorp', '', '']]
        )
    ])
    def test_providers_tab(self, url, expected_providers):
        """Ensure providers tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for providers tab button to be visible
        providers_tab_button = self.wait_for_element(By.ID, 'module-tab-link-providers')

        # Ensure the providers tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-providers', ensure_displayed=False).is_displayed() == False

        # Click on providers tab
        providers_tab_button.click()

        # Obtain tab content
        providers_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-providers')

        # Ensure tab is displayed
        self.assert_equals(lambda: providers_tab_content.is_displayed(), True)

        providers_table = providers_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = providers_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_providers) + 1)

        for row_itx, expected_row in enumerate(
                [['Name', 'Namespace', 'Source', 'Version']] +
                expected_providers):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 4 columns
            assert len(row_columns) == 4

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    def test_integrations_tab(self):
        """Ensure integrations tab is displayed correctly."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

        # Ensure tab is displayed
        self.assert_equals(lambda: integrations_tab_content.is_displayed(), True)

        integrations_table = integrations_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = integrations_table.find_elements(By.TAG_NAME, 'tr')

        expected_integrations = [
            [
                'Create module version using source archive',
                f'POST {self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/${version}/upload", https=True)}\n' +
                'Source ZIP file must be provided as data.'
            ],
            [
                'Trigger module version import',
                f'POST {self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/${version}/import", https=True)}'
            ],
            [
                'Bitbucket hook trigger',
                f'{self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/hooks/bitbucket", https=True)}'
            ],
            [
                'Github hook trigger (Coming soon)',
                f'{self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/hooks/github", https=True)}'
            ],
            [
                'Gitlab hook trigger (Coming soon)',
                f'{self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/hooks/gitlab", https=True)}'
            ],
            [
                'Mark module version as published',
                f'POST {self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/${version}/publish", https=True)}'
            ]
        ]

        # Check number of rows in tab
        assert len(table_rows) == len(expected_integrations)

        for row_itx, expected_row in enumerate(expected_integrations):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'td')

            ## Ensure each table row has 4 columns
            assert len(row_columns) == 2

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    def test_integration_tab_index_version(self):
        """Test indexing a new module version from the integration tab"""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

        # Ensure publish button exists and is not checked
        assert integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').is_selected() == False

        # Type version number and submit form
        integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('5.2.1')
        integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

        # Wait for success message to be displayed
        success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
        self.assert_equals(lambda: success_message.is_displayed(), True)
        self.assert_equals(lambda: success_message.text, 'Successfully indexed version')

        # Check error message is not displayed
        error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
        assert error_message.is_displayed() == False

        # Ensure version create endpoint was called and publish was not
        self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='5.2.1')
        self._api_version_publish_mock.assert_not_called()

    def test_integration_tab_index_version_and_publish(self):
        """Test indexing and publishing a new module version from the integration tab"""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

        # Ensure publish button exists and is not checked
        publish_checkbox = integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish')
        assert publish_checkbox.is_selected() == False

        # Check publish checkbox
        publish_checkbox.click()

        # Type version number and submit form
        integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('5.2.1')
        integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

        # Wait for success message to be displayed
        success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
        self.assert_equals(lambda: success_message.is_displayed(), True)
        self.assert_equals(lambda: success_message.text, 'Successfully indexed and published version.')

        # Check error message is not displayed
        error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
        assert error_message.is_displayed() == False

        # Ensure version create and publish endpoints were called
        self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='5.2.1')
        self._api_version_publish_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='5.2.1')

    def test_integration_tab_index_version_with_publish_disabled(self):
        """Test indexing a new module version from the integration tab whilst publishing is not possible"""
        with self.update_mock(self._config_publish_api_keys_mock, 'new', ['abcdefg']):
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Wait for integrations tab button to be visible
            integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

            # Ensure the integrations tab content is not visible
            assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

            # Click on integrations tab
            integrations_tab_button.click()

            integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

            # Ensure publish button exists and is not disaplyed
            assert integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').is_displayed() == False

            # Ensure publish button container is not displayed
            assert integrations_tab_content.find_element(By.ID, 'integrations-index-module-version-publish').is_displayed() == False

            # Type version number and submit form
            integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('2.2.2')
            integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

            # Wait for success message to be displayed
            success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
            self.assert_equals(lambda: success_message.is_displayed(), True)
            self.assert_equals(lambda: success_message.text, 'Successfully indexed version')

            # Check error message is not displayed
            error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
            assert error_message.is_displayed() == False

            # Ensure version create endpoint was called and publish was not
            self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')
            self._api_version_publish_mock.assert_not_called()

    def test_integration_tab_index_version_with_indexing_failure(self):
        """Test indexing a new module version from the integration tab with an indexing failure"""
        # Update indexing mocks to cause indexing failure
        with self.update_mock(
                self._api_version_create_mock,
                'return_value',
                ({'status': 'Error', 'message': 'Unittest error message'}, 500)):
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Wait for integrations tab button to be visible
            integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

            # Ensure the integrations tab content is not visible
            assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

            # Click on integrations tab
            integrations_tab_button.click()

            integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

            # Type version number and submit form
            integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('2.2.2')
            integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

            # Wait for error message to be displayed
            error_message = self.wait_for_element(By.ID, 'index-version-error', parent=integrations_tab_content)
            assert error_message.text == 'Unittest error message'

            success_message = integrations_tab_content.find_element(By.ID, 'index-version-success')
            assert success_message.is_displayed() == False

            # Ensure version create endpoint was called and publish was not
            self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')
            self._api_version_publish_mock.assert_not_called()

    def test_integration_tab_index_version_with_publishing_failure(self):
        """Test indexing a new module version from the integration tab with a publishing failure"""
        # Update mock to cause publishing failure
        with self.update_mock(
                self._api_version_publish_mock,
                'return_value',
                ({'status': 'Error', 'message': 'Unittest publish error message'}, 500)):
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Wait for integrations tab button to be visible
            integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

            # Ensure the integrations tab content is not visible
            assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

            # Click on integrations tab
            integrations_tab_button.click()

            integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

            # Type version number
            integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('2.2.2')
            # Check publish button
            integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').click()
            # Click indexing button
            integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

            # Wait for error message to be displayed
            error_message = self.wait_for_element(By.ID, 'index-version-error', parent=integrations_tab_content)
            assert error_message.text == 'Unittest publish error message'

            success_message = integrations_tab_content.find_element(By.ID, 'index-version-success')
            assert success_message.is_displayed() == True
            assert success_message.text == 'Successfully indexed version'

            # Ensure version create endpoint was called and publish was not
            self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')
            self._api_version_publish_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')

    @pytest.mark.parametrize('current_version,expected_versions,expected_selected_version', [
        # On root page without version
        (None, ['1.5.0 (latest)', '1.2.0'], '1.5.0 (latest)'),
        # On latest version
        ('1.5.0', ['1.5.0 (latest)', '1.2.0'], '1.5.0 (latest)'),
        # On previous version
        ('1.2.0', ['1.5.0 (latest)', '1.2.0'], '1.2.0'),
        # On beta version
        ('1.6.1-beta', ['1.5.0 (latest)', '1.2.0', '1.6.1-beta (beta)'], '1.6.1-beta (beta)'),
        # On unpublished version
        ('1.6.0', ['1.5.0 (latest)', '1.2.0', '1.6.0 (unpublished)'], '1.6.0 (unpublished)'),
        # On beta unpublished version
        ('1.0.0-beta', ['1.5.0 (latest)', '1.2.0', '1.0.0-beta (beta) (unpublished)'], '1.0.0-beta (beta) (unpublished)')
    ])
    def test_version_dropdown(self, current_version, expected_versions, expected_selected_version):
        """Test version dropdown contains expected values."""
        # Go to page
        url = self.get_url('/modules/moduledetails/fullypopulated/testprovider')
        if current_version:
            url += f'/{current_version}'

        self.selenium_instance.get(url)

        select_dropdown = self.wait_for_element(By.ID, 'version-select')
        version_options = select_dropdown.find_elements(By.TAG_NAME, 'option')

        # Check the number of items
        #assert len(version_options) == len(expected_versions)

        # Ensure the current selected item is as expected
        assert expected_selected_version == Select(select_dropdown).first_selected_option.text

        # Check each of the select options
        for itx, version_item in enumerate(version_options):
            assert version_item.text == expected_versions[itx]

    def test_example_file_contents(self):
        """Check example files are displayed correctly."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Select example from dropdown
        select = Select(self.wait_for_element(By.ID, 'example-select'))
        select.select_by_visible_text('examples/test-example')

        # Wait for page to reload and example title to be displayed
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'current-submodule').text,
            'Example: examples/test-example'
        )

        # Ensure example files tab is displayed
        self.wait_for_element(By.ID, 'module-tab-link-example-files').click()

        file_tab_content = self.wait_for_element(By.ID, 'module-tab-example-files')

        # Ensure all files are displayed in file list
        file_list = file_tab_content.find_element(By.ID, 'example-file-list-nav').find_elements(By.TAG_NAME, 'a')

        # Ensure files match expected order and name
        assert [file.text for file in file_list] == ['main.tf', 'data.tf', 'variables.tf']

        # Ensure contents of main.tf is shown in data
        expected_main_tf_content = f'# Call root module\nmodule "root" {{\n  source  = "localhost:{self.SERVER.port}/moduledetails/fullypopulated/testprovider"\n  version = "1.5.0"\n}}'
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == expected_main_tf_content

        # Select main.tf file and check content
        file_list[0].click()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == expected_main_tf_content

        # Select data.tf and check content
        file_list[1].click()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == '# This contains data objects'

        # Select variables.tf and check content
        file_list[2].click()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == 'variable "test" {\n  description = "test variable"\n  type = string\n}'

    def test_delete_module_version(self):
        """Check provider logos are displayed correctly."""

        self.perform_admin_authentication(password='unittest-password')

        # Create test module version
        namespace = Namespace(name='moduledetails')
        module = Module(namespace=namespace, name='fullypopulated')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        # Create test module version
        module_version = ModuleVersion(module_provider=module_provider, version='2.5.5')
        module_version.prepare_module()
        module_version.publish()

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/2.5.5'))

        # Click on settings tab
        tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
        tab.click()

        # Ensure the verification text is not shown

        # Click on the delete module version button
        delete_button = self.selenium_instance.find_element(By.ID, 'module-version-delete-button')
        delete_button.click()

        # Ensure the verification text is shown
        verification_div = self.selenium_instance.find_element(By.ID, 'confirm-delete-module-version-div')
        assert 'Confirm deletion of module version 2.5.5:' in verification_div.text

        # Click checkbox for verifying deletion
        delete_checkbox = verification_div.find_element(By.ID, 'confirm-delete-module-version')
        delete_checkbox.click()

        # Click delete module version button again
        delete_button.click()

        # Ensure user is redirected to module page
        assert self.selenium_instance.current_url == self.get_url('/modules/moduledetails/fullypopulated/testprovider')

        # Ensure module version no longer exists
        assert ModuleVersion.get(module_provider=module_provider, version='2.5.5') is None

    @pytest.mark.parametrize('allow_custom_git_url_setting', [True, False])
    def test_git_provider_config(self, allow_custom_git_url_setting):
        """Ensure git provider configuration work as expected."""

        with self.update_mock(self._config_allow_custom_repo_urls_module_provider, 'new', allow_custom_git_url_setting):

            self.perform_admin_authentication(password='unittest-password')

            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/2.5.5'))

            # Click on settings tab
            tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
            tab.click()

            # Check git provider dropdown
            git_provider_select_element = self.selenium_instance.find_element(By.ID, 'settings-git-provider')

            expected_git_providers = ['testgitprovider', 'repo_url_tests', 'repo_url_tests_uri_encoded']
            if allow_custom_git_url_setting:
                expected_git_providers.insert(0, 'Custom')

            for option in git_provider_select_element.find_elements(By.TAG_NAME, 'option'):
                # Check option name matches expected
                expected_name = expected_git_providers.pop(0)
                assert option.text == expected_name

                # Ensure git provider pk is used for value of option, or
                # custom option has an empty value
                expected_value = ''
                if expected_name != 'Custom':
                    git_provider = GitProvider.get_by_name(expected_name)
                    expected_value = str(git_provider.pk)
                assert option.get_attribute('value') == expected_value

            git_provider_select = Select(git_provider_select_element)

            if allow_custom_git_url_setting:
                # Ensure the currently selected item is custom
                assert git_provider_select_element.get_attribute('value') == ''

            # Select a different git provider and save
            git_provider_select.select_by_visible_text('testgitprovider')

            try:
                # Press Update button
                self.selenium_instance.find_element(By.ID, 'module-provider-settings-update').click()

                self.assert_equals(lambda: self.wait_for_element(By.ID, 'settings-status-success').text, 'Settings Updated')

                module_provider = ModuleProvider(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
                assert module_provider._get_db_row()['git_provider_id'] == 1

                # Reload page, assert the new git provider has been set
                self.selenium_instance.refresh()
                git_provider_select_element = self.selenium_instance.find_element(By.ID, 'settings-git-provider')
                self.assert_equals(lambda: git_provider_select_element.get_attribute('value'), '1')

                # If custom git urls is enabled, reset back to custom and save
                if allow_custom_git_url_setting:
                    git_provider_select = Select(git_provider_select_element)
                    git_provider_select.select_by_visible_text('Custom')
                    self.selenium_instance.find_element(By.ID, 'module-provider-settings-update').click()
                    # Ensure the DB row is set to custom
                    module_provider._cache_db_row = None
                    assert module_provider._get_db_row()['git_provider_id'] == None

            finally:
                # Reset git provider for module
                ModuleProvider(
                    Module(
                        Namespace('moduledetails'),
                        'fullypopulated'),
                    'testprovider'
                ).update_attributes(git_provider_id=None)
