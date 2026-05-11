
from dataclasses import dataclass
from typing import List, Dict, Optional
import json

@dataclass(frozen=True)
class ModuleSecurityResultLocation:
    filename: str
    start_line: int
    end_line: int

    @classmethod
    def from_dict(self, data: Dict[str, any]) -> 'ModuleSecurityResultLocation':
        """Return dict representation"""
        return ModuleSecurityResultLocation(
            filename=data.get("filename"),
            start_line=data.get("start_line"),
            end_line=data.get("end_line"),
        )

    def to_dict(self) -> Dict[str, any]:
        """Return dict representation"""
        return {
            "filename": self.filename,
            "start_line": self.start_line,
            "end_line": self.end_line,
        }

@dataclass(frozen=True)
class ModuleSecurityResult:

    rule_id: str
    long_id: str
    rule_description: str
    provider: str
    service: str
    impact: str
    resolution: str
    links: list[str]
    description: str
    severity: str
    warning: bool
    status: bool
    resource: str
    location: ModuleSecurityResultLocation

    def to_dict(self) -> str:
        """To common dict"""
        return {
            'rule_id': self.rule_id,
            'long_id': self.long_id,
            'rule_description': self.rule_description,
            'rule_provider': self.provider,
            'rule_service': self.service,
            'impact': self.impact,
            'resolution': self.resolution,
            'links': self.links,
            'description': self.description,
            'severity': self.severity,
            'warning': self.warning,
            'status': self.status,
            'resource': self.resource,
            'location': self.location.to_dict()
        }

class TfsecModuleSecurityResult(ModuleSecurityResult):

    @classmethod
    def from_data(cls, data: Dict[str, any]) -> 'TfsecModuleSecurityResult':
        return cls(
            rule_id=data.get("rule_id"),
            long_id=data.get("long_id"),
            rule_description=data.get("rule_description"),
            provider=data.get("rule_provider"),
            service=data.get("rule_service"),
            impact=data.get("impact"),
            resolution=data.get("resolution"),
            links=data.get("links"),
            description=data.get("description"),
            severity=data.get("severity"),
            warning=data.get("warning"),
            status=data.get("status"),
            resource=data.get("resource"),
            location=ModuleSecurityResultLocation.from_dict(data.get("location")),
        )


class TrivyModuleSecurityResult(ModuleSecurityResult):
    
    @classmethod
    def from_data(cls, data: Dict[str, any], target: str) -> 'TrivyModuleSecurityResult':
        """Handle misconfiguration"""
        cause_metadata = data.get("CauseMetadata", {})
        return cls(
            rule_id=data.get("ID"),
            long_id=data.get("Query"),
            rule_description=data.get("Title"),
            provider=cause_metadata.get("Provider", None),
            service=cause_metadata.get("Service"),
            impact=data.get("Description"),
            resolution=data.get("Resolution"),
            links=data.get("References"),
            description=data.get("Message"),
            severity=data.get("Severity"),
            warning=False,
            status=1 if data.get("Status") == "PASS" else 0,
            resource=cause_metadata.get("Resource"),
            location=ModuleSecurityResultLocation(
                filename=target,
                start_line=cause_metadata.get("StartLine"),
                end_line=cause_metadata.get("EndLine"),
            ),
        )


@dataclass(frozen=True)
class ModuleSecurityResults:
    results: List[ModuleSecurityResult]

    @classmethod
    def from_tfsec_data(cls, results: Dict[any, any]) -> 'ModuleSecurityResults':
        """Interpret tfsec data and convert to common result data"""
        result_objects = []
        for result in results.get("results", []):
            result_obj = TfsecModuleSecurityResult.from_data(result)
            if result_obj:
                result_objects.append(result_obj)
        return ModuleSecurityResults(results=result_objects)

    @classmethod
    def from_trivy_result(cls, result: Dict[any, any]):
        """Convert trivy result to results per misconfiguration"""
        result_objects = []
        target = result.get("Target")
        for misconfig in result.get("Misconfigurations", []):
            result_obj = TrivyModuleSecurityResult.from_data(data=misconfig, target=target)
            if result_obj:
                result_objects.append(result_obj)
        return result_objects

    @classmethod
    def from_trivy_data(cls, results: Dict[any, any]) -> 'ModuleSecurityResults':
        """Interpret trivy data and convert to common result data"""
        result_objects = []
        for result in results.get("Results", []):
            result_objects += cls.from_trivy_result(result=result)
        return ModuleSecurityResults(results=result_objects)

    @classmethod
    def from_json(cls, data_str: str) -> Optional['ModuleSecurityResults']:
        # Attempt to convert from JSON
        try:
            data = json.loads(data_str)
        except json.JSONDecodeError as exc:
            print("Failed to decode module security JSON data: " + str(exc))
            return None

        if not isinstance(data, dict):
            print("Module security data is not a valid dict: " + str(type(data)))
            return None

        # Determine if data appears to be trivy
        if data.get("SchemaVersion") == 2:
            return cls.from_trivy_data(data)
        elif isinstance(data.get("results"), list):
            return cls.from_tfsec_data(data)
    
        print("Unable to determine schema of module security data")
        return None


    def to_dict(self) -> Dict[str, any]:
        """Convert data to common JSON"""
        return {
            "results": [
                result.to_dict()
                for result in self.results
            ]
        }
