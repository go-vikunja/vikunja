# Contributing to Vikunja MCP Server

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Development Setup

### Prerequisites
- Node.js 20 or higher
- Redis server (for rate limiting)
- Vikunja instance (for testing)
- Git

### Getting Started

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/vikunja-mcp-server.git
   cd vikunja-mcp-server
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your Vikunja URL and test token
   ```

4. **Run tests**
   ```bash
   npm test
   ```

5. **Start development server**
   ```bash
   npm run dev
   ```

## Development Workflow

### Before You Start

1. Check existing issues and pull requests
2. Create an issue to discuss major changes
3. Fork the repository
4. Create a feature branch: `git checkout -b feature/your-feature-name`

### While Developing

1. **Write tests first** - Follow TDD approach
2. **Maintain code quality**:
   ```bash
   npm run lint     # Check linting
   npm run format   # Format code
   npm test         # Run tests
   npm run test:coverage  # Check coverage
   ```

3. **Follow TypeScript best practices**:
   - Use strict type checking
   - Avoid `any` types
   - Document complex functions

4. **Keep commits atomic and well-described**:
   ```bash
   git commit -m "feat: add bulk_delete_tasks tool"
   git commit -m "fix: handle rate limit errors correctly"
   git commit -m "docs: update API reference"
   ```

### Code Standards

#### Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `chore:` - Build/tooling changes

#### Code Style
- Use Prettier for formatting (runs automatically)
- Follow ESLint rules
- Use meaningful variable names
- Add JSDoc comments for public APIs

#### Testing Requirements
- Unit tests for all new functions
- Integration tests for new tools
- Maintain >90% code coverage
- All tests must pass before merging

### Submitting Changes

1. **Ensure all tests pass**
   ```bash
   npm test
   npm run test:coverage
   ```

2. **Ensure linting passes**
   ```bash
   npm run lint
   ```

3. **Update documentation** if needed:
   - README.md for new features
   - docs/API.md for new tools
   - docs/EXAMPLES.md for new use cases
   - CHANGELOG.md for notable changes

4. **Push and create pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Fill out the PR template**:
   - Describe what changed and why
   - Link related issues
   - Add screenshots/examples if applicable
   - Check all boxes in the checklist

## What to Contribute

### Good First Issues
- Additional error messages
- Documentation improvements
- Test coverage improvements
- Example workflows
- Integration guides for new platforms

### Feature Contributions
- New MCP tools (must match Vikunja API capabilities)
- Performance optimizations
- Additional bulk operations
- Webhook support
- Metrics and monitoring

### Bug Fixes
- Check existing issues
- Add regression tests
- Document the fix

## Project Structure

```
vikunja-mcp-server/
├── src/
│   ├── config/          # Configuration management
│   ├── auth/            # Authentication logic
│   ├── ratelimit/       # Rate limiting implementation
│   ├── vikunja/         # Vikunja API client
│   ├── tools/           # MCP tool implementations
│   ├── resources/       # MCP resource providers
│   ├── utils/           # Shared utilities
│   └── index.ts         # Server entry point
├── tests/
│   ├── unit/            # Unit tests (mirror src/)
│   └── integration/     # Integration tests
├── docs/                # User documentation
└── [config files]       # TypeScript, ESLint, etc.
```

## Adding a New Tool

1. **Define the tool** in `src/tools/[category].ts`:
   ```typescript
   export const MyToolSchema = z.object({
     field: z.string(),
   });
   
   export class MyTools {
     async myTool(args: z.infer<typeof MyToolSchema>): Promise<Result> {
       // Implementation
     }
   }
   ```

2. **Register in registry** (`src/tools/registry.ts`):
   ```typescript
   this.registerTool(
     'my_tool',
     'Description of what it does',
     MyToolSchema,
     async (args, ctx) => this.myTools.myTool(args, ctx)
   );
   ```

3. **Add tests** (`tests/unit/tools/[category].test.ts`):
   ```typescript
   describe('myTool', () => {
     it('should do something', async () => {
       // Test implementation
     });
   });
   ```

4. **Update documentation**:
   - Add to docs/API.md
   - Add example to docs/EXAMPLES.md
   - Update README.md tool count

## Testing Guidelines

### Unit Tests
- Mock external dependencies (Vikunja API, Redis)
- Test success and error cases
- Test input validation
- Test rate limiting

### Integration Tests
- Test actual MCP protocol flow
- Test tool registration
- Test error propagation

### Coverage Requirements
- Minimum 90% overall coverage
- 100% for critical paths (auth, rate limiting)
- All new code must include tests

## Documentation Guidelines

### API Documentation
- Include input/output schemas
- Provide examples
- Document error cases
- Use TypeScript types

### User Documentation
- Write clear, concise explanations
- Include practical examples
- Show complete workflows
- Link related documentation

## Questions?

- Open an issue for questions
- Join Vikunja Discord for discussions
- Check existing documentation first

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Vikunja MCP Server! 🚀
