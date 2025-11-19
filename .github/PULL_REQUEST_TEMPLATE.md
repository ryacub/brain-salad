## Description

Please provide a clear and concise description of your changes.

## Related Issue

Closes #(issue number)

If there is no related issue, please explain the motivation for this PR.

## Type of Change

Please check the relevant option(s):

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Code refactoring (no functional changes)
- [ ] Performance improvement
- [ ] Test coverage improvement
- [ ] Build/CI configuration change

## Testing Checklist

Please confirm that you have completed the following:

### Code Quality
- [ ] `go test ./...` - All tests pass
- [ ] `golangci-lint run` - No lint warnings (or `go vet ./...` for basic checks)
- [ ] `go fmt ./...` - Code is properly formatted
- [ ] `go build ./...` - Build succeeds

### Test Coverage
- [ ] Added new tests for new functionality (if applicable)
- [ ] All existing tests still pass
- [ ] Manual testing performed (describe below if applicable)

### Manual Testing Checklist
- [ ] CLI commands work as expected
- [ ] Web server starts successfully (if applicable)
- [ ] Database operations work correctly
- [ ] LLM integration functions properly
- [ ] Configuration changes applied successfully

**Manual Testing Performed:**

*(Provide a summary of manual testing performed. Examples:)*
- Tested `tm dump "test idea"` - ✅ Successfully captured and analyzed idea
- Verified `tm review` functionality - ✅ Filtered and displayed ideas correctly
- Confirmed LLM provider switching - ✅ Claude, OpenAI, and rule-based working
- Tested web interface at `localhost:8080` - ✅ Dashboard loads properly
- Validated configuration changes - ✅ Settings applied correctly

## Documentation

- [ ] Updated relevant documentation (README, docs/, code comments)
- [ ] Updated API documentation (if applicable)
- [ ] Added/updated examples (if applicable)
- [ ] No documentation updates needed

## Breaking Changes

If this PR introduces breaking changes, please describe:
- What breaks
- Migration path for existing users
- Version bump required (major/minor/patch)

## Screenshots/Output

If applicable, add screenshots or console output to demonstrate the changes.

```
Paste relevant output here
```

## Additional Notes

Any additional information that reviewers should know about this PR.

## Checklist

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] My changes generate no new warnings
- [ ] I have updated the CHANGELOG (if applicable)
