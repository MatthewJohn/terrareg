
import unittest.mock

import pytest

from test.unit.terrareg import TerraregUnitTest
import terrareg.module_version_create


class TestModuleVersionCreate(TerraregUnitTest):
    """Test module_version_create context manager"""

    def test_success(self):
        """Test successfully running module_version_create"""

        test_module_version = unittest.mock.MagicMock()
        test_module_version.prepare_module = unittest.mock.MagicMock()
        test_module_version.finalise_module = unittest.mock.MagicMock()
        test_module_version.delete = unittest.mock.MagicMock()

        inner_called = False
        with terrareg.module_version_create.module_version_create(test_module_version):
            inner_called = True
        
        test_module_version.prepare_module.assert_called_once_with()
        test_module_version.finalise_module.assert_called_once_with()
        test_module_version.delete.assert_not_called()
        assert inner_called == True

    def test_inner_failure(self):
        """Test with failure case of code within wrapped code"""

        test_module_version = unittest.mock.MagicMock()
        test_module_version.prepare_module = unittest.mock.MagicMock()
        test_module_version.finalise_module = unittest.mock.MagicMock()
        test_module_version.delete = unittest.mock.MagicMock()

        class UnittestException(Exception):
            pass

        with pytest.raises(UnittestException):
            with terrareg.module_version_create.module_version_create(test_module_version):
                raise UnittestException()
        
        test_module_version.prepare_module.assert_called_once_with()
        test_module_version.finalise_module.assert_not_called()
        test_module_version.delete.assert_called_once_with()

    def test_prepare_module_failure(self):
        """Test running module_version_create with prepare_module exception"""

        class UnittestException(Exception):
            pass

        def raise_exception():
            raise UnittestException

        test_module_version = unittest.mock.MagicMock()
        test_module_version.prepare_module = unittest.mock.MagicMock()
        test_module_version.prepare_module.side_effect = raise_exception
        test_module_version.finalise_module = unittest.mock.MagicMock()
        test_module_version.delete = unittest.mock.MagicMock()

        inner_called = False
        with pytest.raises(UnittestException):
            with terrareg.module_version_create.module_version_create(test_module_version):
                inner_called = True
        
        test_module_version.prepare_module.assert_called_once_with()
        test_module_version.finalise_module.assert_not_called()

        # Delete should not be called, as prepare_module inserts the row into the database
        test_module_version.delete.assert_not_called()

        # Wrapped code should not be called
        assert inner_called == False

    def test_finalise_module_failure(self):
        """Test module_version_create with finalise_module exception"""

        class UnittestException(Exception):
            pass

        def raise_exception():
            raise UnittestException

        test_module_version = unittest.mock.MagicMock()
        test_module_version.prepare_module = unittest.mock.MagicMock()
        test_module_version.finalise_module = unittest.mock.MagicMock()
        test_module_version.finalise_module.side_effect = raise_exception
        test_module_version.delete = unittest.mock.MagicMock()

        inner_called = False
        with pytest.raises(UnittestException):
            with terrareg.module_version_create.module_version_create(test_module_version):
                inner_called = True
        
        test_module_version.prepare_module.assert_called_once_with()
        test_module_version.finalise_module.assert_called_once_with()
        test_module_version.delete.assert_called_once_with()

        assert inner_called == True
