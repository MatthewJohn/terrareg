
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
        assert module_version.module_details.get_graph_json() == {
            "nodes": [
                {"data": {"child_count": 0, "id": "aws_s3_bucket.test_bucket", "label": "aws_s3_bucket.test_bucket", "parent": "root"}, "style": {}},
                {"data": {"child_count": 0, "id": "aws_s3_object.test_obj_root_module", "label": "aws_s3_object.test_obj_root_module", "parent": "root"}, "style": {}},
                {"data": {"child_count": 0, "id": "module.submodule-call.aws_ec2_instance.test_instance", "label": "aws_ec2_instance.test_instance", "parent": "module.submodule-call"}, "style": {}},
                {"data": {"child_count": 1, "id": "module.submodule-call", "label": "submodule-call"}, "style": {"background-color": "#F8F7F9", "color": "#000000", "font-weight": "bold", "text-valign": "top"}},
                {"data": {"child_count": 2, "id": "root", "label": "Root Module"}, "style": {"background-color": "#F8F7F9", "color": "#000000", "font-weight": "bold", "text-valign": "top"}}
            ],
            "edges": [
                {"classes": ["module.submodule-call-root"], "data": {"id": "root.module.submodule-call", "source": "module.submodule-call", "target": "root"}}
            ],
        }

        assert module_version.get_submodules()[0].module_details.get_graph_json() == {
            "nodes": [
                {"data": {"child_count": 0, "id": "aws_ec2_instance.test_instance", "label": "aws_ec2_instance.test_instance", "parent": "root"}, "style": {}},
                {"data": {"child_count": 1, "id": "root", "label": "Root Module"}, "style": {"background-color": "#F8F7F9", "color": "#000000", "font-weight": "bold", "text-valign": "top"}}
            ],
            "edges": [],
        }

        assert module_version.get_examples()[0].module_details.get_graph_json() == {
            "nodes": [
                {"data": {"child_count": 0, "id": "module.main_call.aws_s3_bucket.test_bucket", "label": "aws_s3_bucket.test_bucket", "parent": "module.main_call"}, "style": {}},
                {"data": {"child_count": 0, "id": "module.main_call.aws_s3_object.test_obj_root_module", "label": "aws_s3_object.test_obj_root_module", "parent": "module.main_call"}, "style": {}},
                {"data": {"child_count": 0, "id": "module.main_call.module.submodule-call.aws_ec2_instance.test_instance", "label": "aws_ec2_instance.test_instance", "parent": "module.main_call.module.submodule-call"}, "style": {}},
                {"data": {"child_count": 2, "id": "module.main_call", "label": "main_call"}, "style": {"background-color": "#F8F7F9", "color": "#000000", "font-weight": "bold", "text-valign": "top"}},
                {"data": {"child_count": 1, "id": "module.main_call.module.submodule-call", "label": "submodule-call"}, "style": {"background-color": "#F8F7F9", "color": "#000000", "font-weight": "bold", "text-valign": "top"}},
                {"data": {"child_count": 0, "id": "root", "label": "Root Module"}, "style": {"background-color": "#F8F7F9", "color": "#000000", "font-weight": "bold", "text-valign": "top"}}
            ],
            "edges": [
                {"classes": ["module-module"], "data": {"id": "module.main_call.module.main_call.module.submodule-call", "source": "module.main_call", "target": "module.main_call.module.submodule-call"}},
                {"classes": ["module.main_call-root"], "data": {"id": "root.module.main_call", "source": "module.main_call", "target": "root"}}
            ]
        }
