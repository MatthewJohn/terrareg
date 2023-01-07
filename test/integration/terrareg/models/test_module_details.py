
from datetime import datetime
import json

import pytest
import sqlalchemy

from terrareg.database import Database
from terrareg.models import Example, ExampleFile, Module, ModuleDetails, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest


class TestModuleDetails(TerraregIntegrationTest):

    def test_create(self):
        """Test creating a ModuleDetails row/object"""
        db = Database.get()
        with db.get_engine().connect() as conn:
            highest_existing_id = conn.execute(
                db.module_details.select().order_by(
                    sqlalchemy.desc(db.module_details.c.id)
                )
            ).fetchone()['id']

        # Create new module details object
        module_details = ModuleDetails.create()

        assert isinstance(module_details, ModuleDetails)
        assert module_details.pk == (highest_existing_id + 1)

        assert module_details.readme_content is None
        assert module_details.terraform_docs is None
        assert module_details.tfsec == {'results': None}
        assert module_details.infracost == {}

    def test_update_attributes(self):
        """Test update_attributes method of ModuleDetails"""
        test_reame_content = 'test readme content'
        test_terraform_docs = '{"test": "output"}'
        test_tfsec = '{"results": [{"test_result": 0}]}'
        test_infracost = '{"totalMonthlyCost": "123.321"}'

        # Create new module details object
        module_details = ModuleDetails.create()
        module_details.update_attributes(
            readme_content=test_reame_content,
            terraform_docs=test_terraform_docs,
            tfsec=test_tfsec,
            infracost=test_infracost
        )

        assert module_details.readme_content == Database.encode_blob(test_reame_content)
        assert module_details.terraform_docs == Database.encode_blob(test_terraform_docs)
        assert module_details.tfsec == json.loads(test_tfsec)
        assert module_details.infracost == json.loads(test_infracost)

    def test_delete(self):
        """Test delete method of ModuleDetails"""
        module_details = ModuleDetails.create()
        module_details_id = module_details.pk

        db = Database.get()

        # Ensure the row can be found in the database
        with db.get_engine().connect() as conn:
            res = conn.execute(
                db.module_details.select().where(
                    db.module_details.c.id == module_details_id
                )
            ).fetchone()
        assert res is not None

        # Delete module details
        module_details.delete()

        # Ensure the row is no longer present in DB
        with db.get_engine().connect() as conn:
            res = conn.execute(
                db.module_details.select().where(
                    db.module_details.c.id == module_details_id
                )
            ).fetchone()

        assert res == None

    def test_graph_json(self):
        """Test graph data conversion to JSON"""
        module_version = ModuleVersion.get(ModuleProvider.get(Module(Namespace.get("moduledetails"), "graph-test"), "provider"), "1.0.0")
        assert module_version.module_details.graph_json == ""
