# Claude Guidelines

## Commit Standards

This project follows a consistent commit message format with emoji prefixes and conventional commit structure.

### Format
```
{emoji} {type}({scope}): {description}

{body}

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Emoji Prefixes
- ✨ `feat`: New features or enhancements
- 🔧 `config`: Configuration changes
- 📋 `docs`: Documentation updates
- 🐛 `fix`: Bug fixes
- ♻️ `refactor`: Code refactoring
- 🧪 `test`: Test additions or modifications
- 🚀 `perf`: Performance improvements
- 🔒 `security`: Security-related changes

### Examples
```
✨ feat(ci): add GitHub actions workflow for automated Go testing
🔧 config: expand Claude Code tool permissions for spec workflow
📋 docs: add GitHub workflow tests specification
✨ feat: implement thread-safe game state management system
```

### Body Guidelines
- Include why the change was made, not just what changed
- Keep lines under 72 characters
- Use present tense ("add" not "added")
- Be descriptive but concise

### Required Elements
- All commits must include the Claude Code attribution footer
- Use appropriate emoji and type prefix
- Include scope when applicable (component, feature, etc.)
- Provide meaningful description

## Development Commands

### Testing
```bash
go test ./...
go build ./...
```

### Linting
Check project for specific linting commands in package.json or Makefile.