# Makefile for Relia OSS

.PHONY: help setup test lint check-security clean build install format

help:
	@echo "Relia Developer Guidelines"
	@echo "--------------------------"
	@echo "make setup          Install dependencies"
	@echo "make test           Run full test suite"
	@echo "make lint           Run linting (ruff) and type checking (mypy)"
	@echo "make format         Format code with ruff"
	@echo "make check-security Run security scans (bandit, pip-audit)"
	@echo "make clean          Remove build artifacts and cache"
	@echo "make build          Build python package"
	@echo "make install        Install package locally"

setup:
	poetry install

test:
	poetry run pytest

lint:
	poetry run ruff check .
	poetry run mypy .

format:
	poetry run ruff format .

check-security:
	poetry run bandit -c pyproject.toml -r relia
	poetry run pip-audit

clean:
	rm -rf dist
	rm -rf build
	rm -rf .coverage
	rm -rf .pytest_cache
	rm -rf .mypy_cache
	rm -rf .ruff_cache
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type d -name "*.egg-info" -exec rm -rf {} +

build:
	poetry build

install:
	pip install .
