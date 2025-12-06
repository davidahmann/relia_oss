# Use official Python runtime as a parent image (Pinned SHA for security)
FROM python:3.11-slim@sha256:193fdd0bbcb3d2ae612bd6cc3548d2f7c78d65b549fcaa8af75624c47474444d

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd -m -r relia && \
    chown relia /app

# Install Poetry
RUN pip install poetry==1.7.0
RUN poetry config virtualenvs.create false

# Copy project definition
COPY pyproject.toml poetry.lock ./

# Install dependencies
RUN poetry install --without dev --no-interaction --no-ansi

# Copy the rest of the application
COPY . .

# Install the package itself
RUN pip install .

# Switch to non-root user
USER relia

# Entry point
ENTRYPOINT ["relia"]
CMD ["--help"]
