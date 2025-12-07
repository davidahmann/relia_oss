from relia.models import ReliaConfig


class ConfigLoader:
    def load(self, path: str = ".relia.yaml") -> ReliaConfig:
        # Pydantic Settings handles env vars and file loading.
        # We pass the custom path via env var to our custom source.
        # We use a try/finally block to ensure we don't pollute the env permanently.
        import os

        original_path = os.environ.get("RELIA_CONFIG_PATH")
        os.environ["RELIA_CONFIG_PATH"] = path

        try:
            return ReliaConfig()
        finally:
            if original_path is None:
                del os.environ["RELIA_CONFIG_PATH"]
            else:
                os.environ["RELIA_CONFIG_PATH"] = original_path
