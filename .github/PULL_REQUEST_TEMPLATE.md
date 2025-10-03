## 📋 Summary

<!-- Briefly describe what this PR changes or adds. -->

## 🔗 Related Issues

<!-- Link to any related issues, discussions, or tickets. -->
Closes #

## 🎯 Type of Change

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Configuration change
- [ ] 🏗️ Build system change
- [ ] ♻️ Refactoring (no functional changes)
- [ ] 🧪 Test improvement
- [ ] 🚀 Performance improvement

## 🧪 Testing

<!-- Describe the tests you ran to verify your changes. -->

### Manual Testing
- [ ] I tested the changes locally
- [ ] I verified the build works with `task build`
- [ ] I ran the binary and confirmed expected behavior

### Automated Testing
- [ ] All existing tests pass (`task test`)
- [ ] I added new tests for new functionality
- [ ] Integration tests pass (if applicable)

## 📋 Code Quality Checklist

### Go Code Standards
- [ ] My code follows Go conventions and best practices
- [ ] I ran `go fmt` to format the code
- [ ] I ran `go vet` and addressed any issues
- [ ] I ran `task lint` and addressed any linting issues
- [ ] I added appropriate error handling
- [ ] I added appropriate logging where needed

### Documentation
- [ ] I updated relevant documentation in `docs/`
- [ ] I updated the README if needed
- [ ] I added or updated code comments for complex logic
- [ ] I updated CLI help text if I added/modified commands

### Dependencies
- [ ] I minimized new dependencies
- [ ] I updated `go.mod` and `go.sum` appropriately
- [ ] New dependencies are justified and documented

## 🔍 Security Considerations

- [ ] This change does not introduce security vulnerabilities
- [ ] I handled sensitive data appropriately
- [ ] I validated all user inputs
- [ ] I considered potential attack vectors

## 📱 Compatibility

- [ ] This change is backward compatible
- [ ] If breaking, I documented the migration path
- [ ] This works on all supported platforms (Linux, macOS, Windows)
- [ ] This works with all supported Go versions (1.19, 1.20, 1.21)

## 🚀 Performance Impact

- [ ] This change does not negatively impact performance
- [ ] I considered memory usage implications
- [ ] I considered CPU usage implications
- [ ] Large operations are optimized or async where appropriate

## 📸 Screenshots/Demo

<!-- If applicable, add screenshots or demo output to help explain your changes. -->

## 🗒️ Additional Notes

<!-- Add any extra context, implementation details, or notes for reviewers. -->

### Breaking Changes
<!-- If this is a breaking change, describe what breaks and how users should migrate. -->

### Future Considerations
<!-- Any follow-up work or considerations for future PRs. -->

---

## 🔄 CI/CD Status

The following automated checks must pass:
- [ ] ✅ Tests (Go unit tests across multiple platforms)
- [ ] ✅ Lint (golangci-lint)
- [ ] ✅ Build (cross-platform builds)
- [ ] ✅ Security (CodeQL, dependency scan, secrets scan)
- [ ] ✅ Documentation (if docs changed)

<!-- 
For maintainers:
- Review the code for Go best practices
- Check that error handling is appropriate
- Verify that logging is consistent
- Ensure tests cover the new functionality
- Confirm documentation is updated
-->