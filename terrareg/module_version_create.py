from contextlib import contextmanager


@contextmanager
def module_version_create(module_version):
    """Handle module version creation"""
    module_version.prepare_module()
    try:
        yield
        module_version.finalise_module()
    except:
        module_version.delete()
        raise
