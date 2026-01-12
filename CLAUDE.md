# CLAUDE.md

Configuration for Claude Code when working with jp-go-config package.

## Standards

Use `/ai-common` skill to load development standards and patterns as needed.

## Package Purpose

jp-go-config provides standardized configuration structures for Go projects with:

- Database configuration (DatabaseConfig)
- Redis configuration (RedisConfig)
- HTTP configuration (HTTPConfig)
- Integration with Viper
- Test configuration builders

## Development Guidelines

This is a **shared package** used across multiple projects. Changes must be:

- Backward compatible
- Well-tested
- Generic (not project-specific)
- Documented in examples
