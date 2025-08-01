# Project Guidelines

## Understanding the Codebase
- Read README.md to understand what this service is about

## Code Documentation
- Write comments for all exported functions and structs
- Do NOT use periods (`.`) at the end of comments unless there are several sentences

# Code practices:
- DO NOT use env variables unless it is specified, for dynamic configuring refer to previously read README.md in order to find out how configuration works

## Testing Requirements
- Write tests for all exported functions
- Place tests in separate files using the `*package*_test` naming convention
- Test files should be in the same directory as the code being tested
- If test already uses mock or requires updated or new mock, use `make generate` or add new line into `generate` command in `Makefile` file
- DO NOT write comments in tests unless they explain something that is not self-evident

## Build and Quality Checks
- Run validation commands before completing work:
  - `make build` - compile the project
  - `go test -v -race ./... | grep -E "(FAIL:|RUN.*failed|panic:|error:)"` - run all tests
  - `make lint` - check code quality
  - `make generate | grep -E "(error:|warning:|failed)"` - generate swagger files and mocks if there were updates in definitions
- Fix any issues found by these commands

## Documentation
- Record important decisions and implementation details in `thinking.md`
- Include instructions for any external setup (database, AWS, APIs, etc.)
- Separate new entries with `-----` delimiter
- If there are significant updates regarding what written in `README.md`, add it there

## Git Workflow Restrictions
- DO NOT run any modifying `git` commands (i.e. `git add`, `git commit`, `git push`), only read commands allowed (i.e. `git status`, `git diff`, `git log`)
- DO NOT delete files - notify me if files become redundant

## File Ignoring Guidelines
- DO NOT review or analyze files in the following directories:
  - `docs/**` - documentation files
  - `**/mocks/**` - generated mock files
  - `vendor/**` - external dependencies
- DO NOT review or analyze files matching these patterns:
  - `**/*_mock.go` - mock files
  - `**/*.gen.go` - generated files
  - `*.pem`, `*.key`, `app.env` - data-sensitive files

