---
title: "dev-stack"
description: "A powerful development stack management tool built in Go for streamlined local development automation"
lead: "Streamline your local development with powerful CLI tools and automated service management"
date: 2024-01-01T00:00:00+00:00
lastmod: 2024-01-01T00:00:00+00:00
draft: false
weight: 50
toc: true
---

# Welcome to dev-stack

A powerful CLI tool built in Go that helps you quickly set up, manage, and tear down development environments with consistent configurations across your team.

## What is dev-stack?

**dev-stack** is a modern CLI tool that provides quick setup, Docker integration, configuration management, and built-in monitoring for development environments.

## Quick Start

### Installation

```bash
# Download the latest release
curl -L -o dev-stack "https://github.com/isaacgarza/dev-stack/releases/latest/download/dev-stack-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m)"
chmod +x dev-stack
sudo mv dev-stack /usr/local/bin/
```

### Basic Usage

```bash
# Initialize a new project
dev-stack init go --name my-app

# Start your development stack
dev-stack up
```

## Key Features

- **Project Templates**: Go, Node.js, Python, and full-stack setups
- **Service Management**: Databases, message queues, monitoring tools
- **Health Monitoring**: Built-in health checks and status monitoring
- **Docker Integration**: Seamless container management

## Documentation

- **[Setup & Installation]({{< relref "setup" >}})**
- **[Usage Guide]({{< relref "usage" >}})**
- **[Services Guide]({{< relref "services" >}})**
- **[Configuration]({{< relref "configuration" >}})**
- **[CLI Reference]({{< relref "reference" >}})**
- **[Contributing]({{< relref "contributing" >}})**

## Get Started

1. **[Complete installation guide]({{< relref "setup" >}})**
2. **[Learn basic usage]({{< relref "usage" >}})**
3. **[Explore available services]({{< relref "services" >}})**
