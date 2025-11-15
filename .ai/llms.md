# jp-go-config - AI Documentation

Progressive loading map for AI assistants working with jp-go-config package.

**Entry Point**: This file should be referenced from CLAUDE.md.

## Package Overview

**Purpose**: Standardized configuration structures for Go projects

**Key Features**:

- DatabaseConfig for PostgreSQL connections
- RedisConfig for Redis connections
- HTTPConfig for HTTP server settings
- Viper integration patterns
- Test configuration builders

## Always Load

- `.ai/llms.md` (this file)

## Load for Complex Tasks

- `.ai/memory.md` - Design decisions, gotchas, backward compatibility notes
- `.ai/context.md` - Current changes (if exists and is current)

## Common Standards (Portable Patterns)

**See** `.ai/common/common-llms.md` for the complete list of common standards.

Load these common standards when working on this package:

### Core Go Patterns

- `common/standards/go/constructors.md` - New* constructor functions
- `common/standards/go/type-organization.md` - Interface and type placement
- `common/standards/go/validation.md` - Input validation patterns

### Testing

- `common/standards/testing/bdd-testing.md` - Ginkgo/Gomega patterns
- `common/standards/testing/test-categories.md` - Test organization

### Documentation

- `common/standards/documentation/pattern-documentation.md` - Documentation structure
- `common/standards/documentation/code-references.md` - Code examples

## Project Standards (Package-Specific)

This package has minimal package-specific standards since it IS a standard itself.

Any package-specific patterns should go in `.ai/project-standards/`

## Loading Strategy

| Task Type | Load These Standards |
|-----------|---------------------|
| Adding new config type | constructors.md, validation.md, type-organization.md |
| Writing tests | bdd-testing.md, test-categories.md |
| Documenting configs | pattern-documentation.md, code-references.md |
| Ensuring compatibility | memory.md (for backward compatibility notes) |

## File Organization

```
jp-go-config/
├── CLAUDE.md                   # Entry point
├── .gitignore                  # Ignores context.md, memory.md, tasks/
└── .ai/
    ├── llms.md                 # This file (loading map)
    ├── README.md               # Documentation about .ai setup
    ├── context.md              # Current work (gitignored)
    ├── memory.md               # Stable knowledge (gitignored)
    ├── tasks/                  # Scratchpad (gitignored)
    ├── project-standards/      # Package-specific (if needed)
    └── common -> ~/code/ai-common  # Symlink to shared standards
```

## Key Principles

1. **Backward Compatibility**: Never break existing config types or behavior
2. **Generic Design**: No project-specific config types in this package
3. **Viper Integration**: Config types designed to work with Viper
4. **Test Builders**: Provide test helpers for creating configs
5. **Validation**: Clear validation and sensible defaults

## Related Documentation

- Common standard: `common/standards/go/jp-go-config.md` - How to USE this package
- This is the implementation, that is the usage guide
