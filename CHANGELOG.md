# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0](https://github.com/isaacgarza/dev-stack/compare/dev-stack-v0.1.0...dev-stack-v1.0.0) (2025-10-03)


### ⚠ BREAKING CHANGES

* Legacy scripts removed - use Go CLI commands instead

### Features

* add the inital dev stack framework ([7f39b27](https://github.com/isaacgarza/dev-stack/commit/7f39b271c3b828b301841d1dcdc1b733e44cc3f3))
* complete Phase 2 - migrate to Go CLI with full feature parity ([#15](https://github.com/isaacgarza/dev-stack/issues/15)) ([17f13fa](https://github.com/isaacgarza/dev-stack/commit/17f13fa88b16c747c61a78ea739af50776d8681e))
* implement Phase 2.1 service management commands with Docker integration ([#3](https://github.com/isaacgarza/dev-stack/issues/3)) ([82a5f64](https://github.com/isaacgarza/dev-stack/commit/82a5f64569f361aada815df5845f6a47ddd11fa7))
* improve documentation and make it automated where possible ([#1](https://github.com/isaacgarza/dev-stack/issues/1)) ([19b4a05](https://github.com/isaacgarza/dev-stack/commit/19b4a05a2bd4c4ec5de91b02522ec077c7c666c0))
* migrate from Makefile to Taskfile build system ([#24](https://github.com/isaacgarza/dev-stack/issues/24)) ([27e11f0](https://github.com/isaacgarza/dev-stack/commit/27e11f0102579accef7971aab1dc4beb81251254))
* phase 1 of migration to go ([#2](https://github.com/isaacgarza/dev-stack/issues/2)) ([ec8569a](https://github.com/isaacgarza/dev-stack/commit/ec8569abb61a28c7e7324108932673be1e815928))
* simplify github workflows ([#14](https://github.com/isaacgarza/dev-stack/issues/14)) ([289e381](https://github.com/isaacgarza/dev-stack/commit/289e38134288c753ade3c179dc372a62e7e22d09))


### Bug Fixes

* **ci:** docs site ([#22](https://github.com/isaacgarza/dev-stack/issues/22)) ([801669d](https://github.com/isaacgarza/dev-stack/commit/801669d0cc019644ea0f335dc55cbe301dd21aa4))
* **ci:** docs site ([#23](https://github.com/isaacgarza/dev-stack/issues/23)) ([227ac3d](https://github.com/isaacgarza/dev-stack/commit/227ac3d7c4fff90de419d59cf609c68a54625d57))
* **ci:** docs site action ([#20](https://github.com/isaacgarza/dev-stack/issues/20)) ([2974af6](https://github.com/isaacgarza/dev-stack/commit/2974af67a62e828933f039073406a414a10abc4f))
* **ci:** docs site action ([#21](https://github.com/isaacgarza/dev-stack/issues/21)) ([e69d92b](https://github.com/isaacgarza/dev-stack/commit/e69d92b05055228889b6f00c323854fb8afcb975))
* **ci:** resolve release-please, TruffleHog, Windows tests, and docs site deployment ([#18](https://github.com/isaacgarza/dev-stack/issues/18)) ([f97e6e6](https://github.com/isaacgarza/dev-stack/commit/f97e6e64c778a06b80cd4c8779ef8f2a3b14bb40))
* update .github action versions; linter errors; go version sync ([#10](https://github.com/isaacgarza/dev-stack/issues/10)) ([fc2bea9](https://github.com/isaacgarza/dev-stack/commit/fc2bea96495ef265800342c57eec92ec4c931965))

## [Unreleased]

### Features
- Automated release process with Release-Please
- Conventional commit enforcement
- Centralized release configuration management
- Multi-platform binary builds (Linux, macOS, Windows)
- Docker image automation with multi-architecture support
- Package manager integration framework (Homebrew, Scoop)

### Build System
- Migrated from Makefile to modern Taskfile build system
- Enhanced build system with intelligent caching and cross-platform support
- Automated configuration file generation
- Git hooks setup for conventional commits

### Documentation
- Comprehensive release process documentation
- Conventional commit guidelines
- Package manager setup instructions

---

*This changelog is automatically maintained by [Release Please](https://github.com/googleapis/release-please).*
