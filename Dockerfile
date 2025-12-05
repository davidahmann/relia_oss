# Use official Python runtime as a parent image
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies (git is often needed for pre-commit or pip deps)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Poetry
RUN pip install poetry==1.7.0
RUN poetry config virtualenvs.create false

# Copy project definition
COPY pyproject.toml poetry.lock ./

# Install dependencies (no dev deps for prod image, but we might want them for CI image)
# For the CLI distribution, we'll install without dev dependencies
RUN poetry install --without dev --no-interaction --no-ansi

# Copy the rest of the application
COPY . .

# Install the package itself
RUN pip install .

# Entry point
ENTRYPOINT ["relia"]
CMD ["--help"]
