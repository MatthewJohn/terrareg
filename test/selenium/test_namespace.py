
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestNamespace(SeleniumTest):
    """Test homepage."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch('terrareg.analytics.AnalyticsEngine.get_total_downloads', return_value=2005))
        cls.register_patch(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        cls.register_patch(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['trustednamespace']))
        super(TestNamespace, cls).setup_class()

    def test_provider_logos(self):
        """Check provider logos are displayed correctly."""
        self.selenium_instance.get(self.get_url('/modules/real_providers'))

        # Ensure all provider logo TOS are displayed
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'provider-tos-aws').text,
            'Amazon Web Services, AWS, the Powered by AWS logo are trademarks of Amazon.com, Inc. or its affiliates.'
        )
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'provider-tos-gcp').text,
            'Google Cloud and the Google Cloud logo are trademarks of Google LLC.'
        )
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'provider-tos-null').text,
            ''
        )
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'provider-tos-datadog').text,
            'All \'Datadog\' modules are designed to work with Datadog. Modules are in no way affiliated with nor endorsed by Datadog Inc.'
        )

        # Check logo for each module
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'real_providers.test-module.aws.1.0.0'
            ).find_element(By.TAG_NAME, 'img').get_attribute('src'),
            self.get_url('/static/images/PB_AWS_logo_RGB_stacked.547f032d90171cdea4dd90c258f47373c5573db5.png')
        )
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'real_providers.test-module.gcp.1.0.0'
            ).find_element(By.TAG_NAME, 'img').get_attribute('src'),
            self.get_url('/static/images/gcp.png')
        )
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'real_providers.test-module.null.1.0.0'
            ).find_element(By.TAG_NAME, 'img').get_attribute('src'),
            self.get_url('/static/images/null.png')
        )
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'real_providers.test-module.datadog.1.0.0'
            ).find_element(By.TAG_NAME, 'img').get_attribute('src'),
            self.get_url('/static/images/dd_logo_v_rgb.png')
        )

        # Ensure no logo is present for unknown provider
        null_module = self.selenium_instance.find_element(
                By.ID, 'real_providers.test-module.doesnotexist.1.0.0')
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            null_module.find_element(By.TAG_NAME, 'img')

    def test_module_details(self):
        """Check that module details are displayed."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails'))

        module = self.wait_for_element(By.ID, 'moduledetails.fullypopulated.testprovider.1.5.0')

        card_title = module.find_element(By.CLASS_NAME, 'module-card-title')
        assert card_title.get_attribute('href') == self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0')
        assert card_title.text == 'moduledetails / fullypopulated'

        # Ensure description and owner is in body
        card_content = module.find_element(By.CLASS_NAME, 'card-content').find_element(By.CLASS_NAME, 'content')

        assert 'This is a test module version for tests.' in card_content.text
        assert 'Owner: This is the owner of the module' in card_content.text

        # Check link to source
        assert module.find_element(By.CLASS_NAME, 'card-source-link').text == 'Source: https://link-to.com/source-code-here'

    def test_verified_module(self):
        """Check that verified modules are displayed."""
        self.selenium_instance.get(self.get_url('/modules/modulesearch'))

        verified_module = self.wait_for_element(By.ID, 'modulesearch.verifiedmodule-oneversion.aws.1.0.0')

        # Check that verified label is displayed
        verified_label = verified_module.find_element(By.CLASS_NAME, 'result-card-label-verified')
        assert verified_label.text == 'unittest verified label'

        # Check non-verified module does not contain the element
        unverified_module = self.wait_for_element(By.ID, 'modulesearch.contributedmodule-oneversion.aws.1.0.0')
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            unverified_module.find_element(By.CLASS_NAME, 'result-card-label-verified')

    def test_trusted_module(self):
        """Check that trusted modules just have trusted label."""
        self.selenium_instance.get(self.get_url('/modules/trustednamespace'))

        trusted_module = self.wait_for_element(By.ID, 'trustednamespace.searchbymodulename4.aws.5.5.5')

        # Check that verified label is displayed
        trusted_label = trusted_module.find_element(By.CLASS_NAME, 'result-card-label-trusted')
        assert trusted_label.text == 'unittest trusted namespace'

        # Ensure that the contributed tag is not shown
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            trusted_module.find_element(By.CLASS_NAME, 'result-card-label-contributed')

    def check_contributed_module(self):
        """Check that contributed module just has contributed label"""
        self.selenium_instance.get(self.get_url('/modules/modulesearch-contributed'))

        contributed_module = self.wait_for_element(By.ID, 'modulesearch-contributed.mixedsearch-result.aws.1.0.0')

        # Check that verified label is displayed
        trusted_label = contributed_module.find_element(By.CLASS_NAME, 'result-card-label-contributed')
        assert trusted_label.text == 'unittest contributed namespace'

        # Ensure that the contributed tag is not shown
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            contributed_module.find_element(By.CLASS_NAME, 'result-card-label-trusted')
