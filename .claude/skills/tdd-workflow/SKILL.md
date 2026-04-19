---
name: tdd-workflow
description: Use this skill when writing new features, fixing bugs, or refactoring code. Enforces test-driven development with 80%+ coverage including unit, integration, and E2E tests.
origin: ECC
---

# Test-Driven Development Workflow

> **Project addendum (opnConfigGenerator):** TDD inner loop is `mise exec -- just test-race` (race detector; flags owned by the `test-race` recipe in `justfile`). Full gate is `mise exec -- just ci-check`. This project follows `AGENTS.md` §Code Quality Policy — zero-tolerance for tech debt; fix failures regardless of when they were introduced. The generic "80% coverage" threshold here is a floor, not a ceiling; prefer load-bearing assertions (test a single mutation on each assertion to prove it fails; see `internal/opnsensegen/deps_isolation_test.go` for a documented example). Use the standard library's `testing` with `stretchr/testify` (`require` for error paths, `assert` for continuation), table-driven tests (see `internal/opnsensegen/template_test.go`), and `t.Parallel()` by default.

This skill ensures all code development follows TDD principles with comprehensive test coverage.

## When to Activate

- Writing new features or functionality
- Fixing bugs or issues
- Refactoring existing code
- Adding API endpoints
- Creating new components

## Core Principles

### 1. Tests BEFORE Code

ALWAYS write tests first, then implement code to make tests pass.

### 2. Coverage Requirements

- Minimum 80% coverage (unit + integration + E2E)
- All edge cases covered
- Error scenarios tested
- Boundary conditions verified

### 3. Test Types

#### Unit Tests

- Individual functions and utilities
- Component logic
- Pure functions
- Helpers and utilities

#### Integration Tests

- API endpoints
- Database operations
- Service interactions
- External API calls

#### E2E Tests (Playwright)

- Critical user flows
- Complete workflows
- Browser automation
- UI interactions

### 4. Git Checkpoints

- If the repository is under Git, create a checkpoint commit after each TDD stage
- Do not squash or rewrite these checkpoint commits until the workflow is complete
- Each checkpoint commit message must describe the stage and the exact evidence captured
- Count only commits created on the current active branch for the current task
- Do not treat commits from other branches, earlier unrelated work, or distant branch history as valid checkpoint evidence
- Before treating a checkpoint as satisfied, verify that the commit is reachable from the current `HEAD` on the active branch and belongs to the current task sequence
- The preferred compact workflow is:
  - one commit for failing test added and RED validated
  - one commit for minimal fix applied and GREEN validated
  - one optional commit for refactor complete
- Separate evidence-only commits are not required if the test commit clearly corresponds to RED and the fix commit clearly corresponds to GREEN

## TDD Workflow Steps

### Step 1: Write User Journeys

```text
As a [role], I want to [action], so that [benefit]

Example:
As a user, I want to search for markets semantically,
so that I can find relevant markets even without exact keywords.
```

### Step 2: Generate Test Cases

For each user journey, create comprehensive test cases:

```typescript
describe('Semantic Search', () => {
  it('returns relevant markets for query', async () => {
    // Test implementation
  })

  it('handles empty query gracefully', async () => {
    // Test edge case
  })

  it('falls back to substring search when Redis unavailable', async () => {
    // Test fallback behavior
  })

  it('sorts results by similarity score', async () => {
    // Test sorting logic
  })
})
```

### Step 3: Run Tests (They Should Fail)

```bash
# For this Go project — see project addendum at the top of this file.
mise exec -- just test-race
# Tests should fail — we haven't implemented yet.
# (Upstream guidance used `npm test`; Go equivalent is go test, wrapped
# by the just recipe so the race detector and CI timeout are applied.)
```

This step is mandatory and is the RED gate for all production changes.

Before modifying business logic or other production code, you must verify a valid RED state via one of these paths:

- Runtime RED:
  - The relevant test target compiles successfully
  - The new or changed test is actually executed
  - The result is RED
- Compile-time RED:
  - The new test newly instantiates, references, or exercises the buggy code path
  - The compile failure is itself the intended RED signal
- In either case, the failure is caused by the intended business-logic bug, undefined behavior, or missing implementation
- The failure is not caused only by unrelated syntax errors, broken test setup, missing dependencies, or unrelated regressions

A test that was only written but not compiled and executed does not count as RED.

Do not edit production code until this RED state is confirmed.

If the repository is under Git, create a checkpoint commit immediately after this stage is validated. Recommended commit message format:

- `test: add reproducer for <feature or bug>`
- This commit may also serve as the RED validation checkpoint if the reproducer was compiled and executed and failed for the intended reason
- Verify that this checkpoint commit is on the current active branch before continuing

### Step 4: Implement Code

Write minimal code to make tests pass:

```typescript
// Implementation guided by tests
export async function searchMarkets(query: string) {
  // Implementation here
}
```

If the repository is under Git, stage the minimal fix now but defer the checkpoint commit until GREEN is validated in Step 5.

### Step 5: Run Tests Again

```bash
# For this Go project — use the project's just recipe.
mise exec -- just test-race
# Tests should now pass.
```

Rerun the same relevant test target after the fix and confirm the previously failing test is now GREEN.

Only after a valid GREEN result may you proceed to refactor.

If the repository is under Git, create a checkpoint commit immediately after GREEN is validated. Recommended commit message format:

- `fix: <feature or bug>`
- The fix commit may also serve as the GREEN validation checkpoint if the same relevant test target was rerun and passed
- Verify that this checkpoint commit is on the current active branch before continuing

### Step 6: Refactor

Improve code quality while keeping tests green:

- Remove duplication
- Improve naming
- Optimize performance
- Enhance readability

If the repository is under Git, create a checkpoint commit immediately after refactoring is complete and tests remain green. Recommended commit message format:

- `refactor: clean up after <feature or bug> implementation`
- Verify that this checkpoint commit is on the current active branch before considering the TDD cycle complete

### Step 7: Verify Coverage

```bash
npm run test:coverage
# Verify 80%+ coverage achieved
```

## Testing Patterns

### Unit Test Pattern (Jest/Vitest)

```typescript
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from './Button'

describe('Button Component', () => {
  it('renders with correct text', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByText('Click me')).toBeInTheDocument()
  })

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn()
    render(<Button onClick={handleClick}>Click</Button>)

    fireEvent.click(screen.getByRole('button'))

    expect(handleClick).toHaveBeenCalledTimes(1)
  })

  it('is disabled when disabled prop is true', () => {
    render(<Button disabled>Click</Button>)
    expect(screen.getByRole('button')).toBeDisabled()
  })
})
```

### API Integration Test Pattern

```typescript
import { NextRequest } from 'next/server'
import { GET } from './route'

describe('GET /api/markets', () => {
  it('returns markets successfully', async () => {
    const request = new NextRequest('http://localhost/api/markets')
    const response = await GET(request)
    const data = await response.json()

    expect(response.status).toBe(200)
    expect(data.success).toBe(true)
    expect(Array.isArray(data.data)).toBe(true)
  })

  it('validates query parameters', async () => {
    const request = new NextRequest('http://localhost/api/markets?limit=invalid')
    const response = await GET(request)

    expect(response.status).toBe(400)
  })

  it('handles database errors gracefully', async () => {
    // Mock database failure
    const request = new NextRequest('http://localhost/api/markets')
    // Test error handling
  })
})
```

### E2E Test Pattern (Playwright)

```typescript
import { test, expect } from '@playwright/test'

test('user can search and filter markets', async ({ page }) => {
  // Navigate to markets page
  await page.goto('/')
  await page.click('a[href="/markets"]')

  // Verify page loaded
  await expect(page.locator('h1')).toContainText('Markets')

  // Search for markets
  await page.fill('input[placeholder="Search markets"]', 'election')

  // Wait for debounce and results
  await page.waitForTimeout(600)

  // Verify search results displayed
  const results = page.locator('[data-testid="market-card"]')
  await expect(results).toHaveCount(5, { timeout: 5000 })

  // Verify results contain search term
  const firstResult = results.first()
  await expect(firstResult).toContainText('election', { ignoreCase: true })

  // Filter by status
  await page.click('button:has-text("Active")')

  // Verify filtered results
  await expect(results).toHaveCount(3)
})

test('user can create a new market', async ({ page }) => {
  // Login first
  await page.goto('/creator-dashboard')

  // Fill market creation form
  await page.fill('input[name="name"]', 'Test Market')
  await page.fill('textarea[name="description"]', 'Test description')
  await page.fill('input[name="endDate"]', '2025-12-31')

  // Submit form
  await page.click('button[type="submit"]')

  // Verify success message
  await expect(page.locator('text=Market created successfully')).toBeVisible()

  // Verify redirect to market page
  await expect(page).toHaveURL(/\/markets\/test-market/)
})
```

## Test File Organization

For this Go project, tests live alongside the code they cover in the same package (conventional Go layout):

```text
internal/
├── opnsensegen/
│   ├── template.go
│   ├── template_test.go             # Unit tests, same package
│   ├── commondevice_test.go         # External-consumer tests (_test pkg)
│   └── deps_isolation_test.go       # //go:build integration
├── generator/
│   ├── vlan.go
│   └── vlan_test.go
└── ...
testdata/
└── base-config.xml                   # Fixtures for tests
```

Integration tests use the `//go:build integration` tag and are run via `mise exec -- just test-integration`. The upstream E2E/Playwright guidance in this skill does NOT apply — this project is a CLI with no UI to drive.

## Mocking External Services

### Supabase Mock

```typescript
jest.mock('@/lib/supabase', () => ({
  supabase: {
    from: jest.fn(() => ({
      select: jest.fn(() => ({
        eq: jest.fn(() => Promise.resolve({
          data: [{ id: 1, name: 'Test Market' }],
          error: null
        }))
      }))
    }))
  }
}))
```

### Redis Mock

```typescript
jest.mock('@/lib/redis', () => ({
  searchMarketsByVector: jest.fn(() => Promise.resolve([
    { slug: 'test-market', similarity_score: 0.95 }
  ])),
  checkRedisHealth: jest.fn(() => Promise.resolve({ connected: true }))
}))
```

### OpenAI Mock

```typescript
jest.mock('@/lib/openai', () => ({
  generateEmbedding: jest.fn(() => Promise.resolve(
    new Array(1536).fill(0.1) // Mock 1536-dim embedding
  ))
}))
```

## Test Coverage Verification

### Run Coverage Report

```bash
npm run test:coverage
```

### Coverage Thresholds

```json
{
  "jest": {
    "coverageThresholds": {
      "global": {
        "branches": 80,
        "functions": 80,
        "lines": 80,
        "statements": 80
      }
    }
  }
}
```

## Common Testing Mistakes to Avoid

### FAIL: WRONG: Testing Implementation Details

```typescript
// Don't test internal state
expect(component.state.count).toBe(5)
```

### PASS: CORRECT: Test User-Visible Behavior

```typescript
// Test what users see
expect(screen.getByText('Count: 5')).toBeInTheDocument()
```

### FAIL: WRONG: Brittle Selectors

```typescript
// Breaks easily
await page.click('.css-class-xyz')
```

### PASS: CORRECT: Semantic Selectors

```typescript
// Resilient to changes
await page.click('button:has-text("Submit")')
await page.click('[data-testid="submit-button"]')
```

### FAIL: WRONG: No Test Isolation

```typescript
// Tests depend on each other
test('creates user', () => { /* ... */ })
test('updates same user', () => { /* depends on previous test */ })
```

### PASS: CORRECT: Independent Tests

```typescript
// Each test sets up its own data
test('creates user', () => {
  const user = createTestUser()
  // Test logic
})

test('updates user', () => {
  const user = createTestUser()
  // Update logic
})
```

## Continuous Testing

### Watch Mode During Development

```bash
npm test -- --watch
# Tests run automatically on file changes
```

### Pre-Commit Hook

```bash
# Runs before every commit
npm test && npm run lint
```

### CI/CD Integration

```yaml
# GitHub Actions
  - name: Run Tests
    run: npm test -- --coverage
  - name: Upload Coverage
    uses: codecov/codecov-action@v3
```

## Best Practices

01. **Write Tests First** - Always TDD
02. **Focus each test on one behavior** - In Go table-driven tests, each row is a distinct behavior; multiple assertions inside one `t.Run` subtest are fine when they all verify the same outcome. Assertions should be load-bearing (mutation-test them where feasible — see `internal/opnsensegen/deps_isolation_test.go` for a documented example). The "one assert per test" rule from upstream doesn't fit the Go convention here.
03. **Descriptive Test Names** - Explain what's tested
04. **Arrange-Act-Assert** - Clear test structure
05. **Mock External Dependencies** - Isolate unit tests
06. **Test Edge Cases** - Null, undefined, empty, large
07. **Test Error Paths** - Not just happy paths
08. **Keep Tests Fast** - Unit tests < 50ms each
09. **Clean Up After Tests** - No side effects
10. **Review Coverage Reports** - Identify gaps

## Success Metrics

- 80%+ code coverage achieved
- All tests passing (green)
- No skipped or disabled tests
- Fast test execution (< 30s for unit tests)
- E2E tests cover critical user flows
- Tests catch bugs before production

---

**Remember**: Tests are not optional. They are the safety net that enables confident refactoring, rapid development, and production reliability.
