---
inclusion: always
---

# Spring Boot Development Guidelines

## Component Structure
- Use @RestController for REST endpoints
- Use @Service for business logic
- Use @Repository for data access
- Use @Component for other beans

## Dependency Injection
- Prefer constructor injection over field injection
- Use final fields for injected dependencies
- Avoid circular dependencies

## API Design
- Follow REST principles for endpoint design
- Use appropriate HTTP methods (GET, POST, PUT, DELETE)
- Return appropriate HTTP status codes
- Use DTOs for request/response objects

## Configuration Management
- Use @ConfigurationProperties for complex configurations
- Externalize configuration using application.yml or application.properties
- Use profiles for environment-specific settings
