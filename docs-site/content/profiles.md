---
title: "Service Profiles"
description: "Predefined service combinations for common development scenarios"
lead: "Quickly start with predefined service combinations"
date: "2025-10-01"
lastmod: "2025-10-11"
draft: false
weight: 35
toc: true
---

<!-- AUTO-GENERATED-START -->
# Service Profiles

Service profiles are predefined combinations of services for common development scenarios. Use them to quickly start your development environment with the right services.

## Using Profiles

```bash
# Start services using a profile
dev-stack up --profile <profile-name>

# List available profiles
dev-stack up --profile <TAB>
```

## Available Profiles


### API Development

Services for API development and testing

**Services included:**

- postgres

- redis

- prometheus


**Quick start:**
```bash
dev-stack up --profile api development
```

---


### Data Engineering

Services for data processing and analytics

**Services included:**

- postgres

- redis

- kafka

- localstack


**Quick start:**
```bash
dev-stack up --profile data engineering
```

---


### Microservices

Full microservices development stack

**Services included:**

- postgres

- redis

- kafka

- jaeger

- prometheus


**Quick start:**
```bash
dev-stack up --profile microservices
```

---


### Minimal Stack

Minimal services for basic development

**Services included:**

- postgres


**Quick start:**
```bash
dev-stack up --profile minimal stack
```

---


### Web Development

Services for web application development

**Services included:**

- postgres

- redis

- jaeger


**Quick start:**
```bash
dev-stack up --profile web development
```

---



## Creating Custom Profiles

You can define custom profiles in your `dev-stack-config.yaml` file:

```yaml
profiles:
  my-profile:
    name: "My Custom Profile"
    description: "Custom services for my project"
    services:
      - postgres
      - redis
      - my-service
```

<!-- AUTO-GENERATED-END -->