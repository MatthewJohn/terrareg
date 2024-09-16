import abc

import terrareg.config


class BaseProduct(abc.ABC):

    @abc.abstractmethod
    def get_tfswitch_product_arg(self) -> str:
        """Return name of tfswitch product argument value"""
        ...

    @abc.abstractmethod
    def get_executable_name(self) -> str:
        """Return executable name for product"""
        ...

class Terraform(BaseProduct):
    """Terraform product"""

    def get_tfswitch_product_arg(self) -> str:
        """Return name of tfswitch product argument value"""
        return "terraform"

    def get_executable_name(self) -> str:
        """Return executable name for product"""
        return "terraform"
    

class OpenTofu(BaseProduct):
    """OpenTofu product"""

    def get_tfswitch_product_arg(self) -> str:
        """Return name of tfswitch product argument value"""
        return "opentofu"

    def get_executable_name(self) -> str:
        """Return executable name for product"""
        return "tofu"


class ProductFactory:

    @staticmethod
    def get_product():
        """Obtain current product"""
        product_enum = terrareg.config.Config().PRODUCT
        if product_enum is terrareg.config.Product.TERRAFORM:
            return Terraform()
        elif product_enum is terrareg.config.Product.OPENTOFU:
            return OpenTofu()
        raise Exception("Could not determine product class")
