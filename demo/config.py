#!/usr/bin/env python3
"""App configuration loader."""

import json
import os

DEFAULTS = {
    "host": "0.0.0.0",
    "port": 8080,
    "debug": False,
    "db_url": "postgresql://localhost:5432/app",
}

def load_config(path=None):
    if path and os.path.exists(path):
        with open(path) as f:
            return {**DEFAULTS, **json.load(f)}
    return DEFAULTS