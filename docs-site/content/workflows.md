---
title: "Workflows"
description: "Common workflows and task sequences for dev-stack"
lead: "Step-by-step guides for common development tasks"
date: 2025-10-07T12:20:14-05:00
lastmod: 2025-10-07T12:20:14-05:00
draft: false
weight: 40
toc: true
---

<!-- AUTO-GENERATED-START -->
# Common Workflows

These workflows guide you through common development tasks using dev-stack.


## Clean Reset

Complete cleanup and fresh start

**Steps:**

1. **Stop services and remove data**
   ```bash
   down --volumes
   ```
   


2. **Clean up all resources**
   ```bash
   cleanup --all
   ```
   


3. **Start fresh services**
   ```bash
   up
   ```
   



---


## Full Stack Development

Complete development environment setup

**Steps:**

1. **Initialize web project**
   ```bash
   init web
   ```
   


2. **Start web development services**
   ```bash
   up --profile web
   ```
   


3. **Monitor all services**
   ```bash
   monitor
   ```
   



---


## Quick Start

Get started with a basic development stack

**Steps:**

1. **Initialize project**
   ```bash
   init
   ```
   


2. **Start database services**
   ```bash
   up postgres redis
   ```
   


3. **Check service status**
   ```bash
   status
   ```
   



---



## Creating Custom Workflows

You can create your own workflows by combining dev-stack commands. Use the `dev-stack workflow` command to see available workflows and execute them interactively.

<!-- AUTO-GENERATED-END -->