
import pytest

from pyvirtualdisplay import Display
import selenium

from . import SeleniumTest

@pytest.fixture(autouse=True, scope="session")
def selenium_fixture(request):
    if not SeleniumTest.RUN_INTERACTIVELY:
        SeleniumTest.display_instance = Display(visible=1, size=SeleniumTest.DEFAULT_RESOLUTION)
        SeleniumTest.display_instance.start()
    SeleniumTest.selenium_instance = selenium.webdriver.Firefox()
    SeleniumTest.selenium_instance.delete_all_cookies()
    SeleniumTest.selenium_instance.implicitly_wait(1)

    def teardown():
        SeleniumTest.selenium_instance.quit()
        if not SeleniumTest.RUN_INTERACTIVELY:
            SeleniumTest.display_instance.stop()

    request.addfinalizer(teardown)