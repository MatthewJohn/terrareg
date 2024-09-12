
from typing import Optional
import os


version_file: str = os.path.join(os.path.dirname(os.path.realpath(__file__)), "version.txt")
VERSION: Optional[str] = None

if os.path.isfile(version_file):
    try:
        with open(version_file, "r") as fh:
            VERSION = fh.readline().strip()
    except Exception as exc:
        print("Failed to read version file", exc)
