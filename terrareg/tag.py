

import json
import re
from dataclasses import dataclass, field
from typing import List, Union, Dict
from enum import Enum

import terrareg.config
from terrareg.errors import InvalidTagsConfigError


class TagType(Enum):

    FLAG = "flag"
    VALUE = "value"


@dataclass
class TagValue:

    label: str
    value: str

    def __post_init__(self):
        """Validate values"""
        if not isinstance(self.value, str):
            raise ValueError("value must be a string")
        if re.match(r"[^a-zA-Z0-9-]", self.value):
            raise ValueError("value must only contain alpha numeric characters and dashes")
        if not len(self.value):
            raise ValueError("value must be a string with a value")
        if not isinstance(self.label, str):
            raise ValueError("label must be a string")
        if not len(self.label):
            raise ValueError("label cannot be empty")
        if re.match(r"[^a-zA-Z0-9 -_]", self.label):
            raise ValueError("label must only contain alph-anumeric, spaces, dashes and underscores")


@dataclass
class Tag:

    id: str
    label: str
    type: TagType
    default_namespaces: List[str] = field(default_factory=list)
    color: str = "yellow"
    allowed_module_provider: bool = True
    allowed_module_version: bool = False
    allowed_in_terrareg_config: bool = False
    values: List['TagValue'] = field(default_factory=list)

    def __post_init__(self):
        """Validate values"""
        if not isinstance(self.id, str):
            raise ValueError("id must be a string")
        if re.match(r"[^a-zA-Z0-9-]", self.id):
            raise ValueError("id must only contain alpha numeric characters and dashes")
        if not len(self.id):
            raise ValueError("id must be a string with a value")
        if not isinstance(self.label, str):
            raise ValueError("label must be a string")
        if not len(self.label):
            raise ValueError("label cannot be empty")
        if re.match(r"[^a-zA-Z0-9 -_]", self.label):
            raise ValueError("label must only contain alph-anumeric, spaces, dashes and underscores")

        if not isinstance(self.default_namespaces, list):
            raise ValueError
        if not all([isinstance(v) for v in self.default_namespaces]):
            raise ValueError
        if not isinstance(self.color, str):
            raise ValueError

        if isinstance(self.type, str):
            try:
                self.type = TagType(self.type.lower())
            except ValueError:
                raise ValueError(f"type must be one of: {', '.join([e.value for e in TagType])}")
        elif not isinstance(self.type, TagType):
            raise ValueError(f"type must be one of: {', '.join([e.value for e in TagType])}")

        if not isinstance(self.allowed_provider, bool):
            raise ValueError("allowed_provider must be a boolean")
        if not isinstance(self.allowed_version, bool):
            raise ValueError("allowed_provider must be a boolean")
        if not isinstance(self.allowed_in_terrareg_config, bool):
            raise ValueError("allowed_in_terrareg_config must be a boolean")
        if self.allowed_in_terrareg_config and not self.allowed_version:
            raise ValueError("Cannot set allowed_in_terrareg_config without allowed_version")
        if not isinstance(self.values, list):
            raise ValueError("values must be a list")

        # Ensure values are set, where appropriate
        if self.type is TagType.FLAG and len(self.values):
            raise ValueError("values cannot be specified for 'flag' type tag")
        elif self.type is TagType.VALUE and not len(self.values):
            raise ValueError("values must be specified for 'value' type tag")
        values: Dict[Union[str, int], TagValue] = {}
        for val_ in self.values:
            if isinstance(dict):
                try:
                    val_ = TagValue(**val_)
                except (ValueError, KeyError) as exc:
                    raise ValueError(f"Invalid configuration in value: {exc}")
            elif not isinstance(val_, TagValue):
                raise ValueError
            if val_.value in values:
                raise ValueError(f"duplicate value found: {val_.value}")
            values[val_.value] = val_
        self.values = values


class TagFactory:

    _INSTANCE: Union['TagFactory', None] = None

    @classmethod
    def get(cls) -> 'TagFactory':
        """"""
        if cls._INSTANCE is None:
            cls._INSTANCE = cls()
        return cls._INSTANCE

    def __init__(self):
        """Initialise tags from config"""
        config = terrareg.config.Config()
        try:
            tags_data = json.loads(config.MODULE_TAGS)
        except json.decoder.JSONDecodeError:
            raise InvalidTagsConfigError("MODULE_TAGS contains invalid JSON")

        if not isinstance(tags_data, list):
            raise InvalidTagsConfigError("MODULE_TAGS must contain a JSON list of objects")

        for itx, tag_data in enumerate(tags_data):
            if not isinstance(tag_data, dict):
                raise InvalidTagsConfigError(f"Invalid module tag in MODULE_TAGS.  tag {itx} must be an object")
            try:
                tag = Tag(**tag_data)
            except (TypeError, ValueError) as exc:
                raise InvalidTagsConfigError(f"Invalid module tag in MODULE_TAGS.  tag {itx}: {str(exc)}")
            if tag.id in self._tags:
                raise InvalidTagsConfigError(f"Invalid module tag in MODULE_TAGS. Duplicate tag id: {tag.id}")
            self._tags = {
                tag.id: tag
            }


