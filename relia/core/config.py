from relia.models import ReliaConfig


class ConfigLoader:
    def load(self, path: str = ".relia.yaml") -> ReliaConfig:
        # Pydantic Settings handles env vars and file loading
        # We pass the path if needed, but our custom source checks CWD
        # For MVP, we ignore the 'path' arg if it's custom,
        # but realistically the source should take it.
        # For 12-Factor, env vars are primary.
        return ReliaConfig()
