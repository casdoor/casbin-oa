# CI Monitoring and Auto-Fix Feature

This document describes the CI monitoring and auto-fix feature implemented in Casbin-OA.

## Overview

This feature automatically monitors CI check failures in Pull Requests and engages GitHub Copilot to fix linter errors. It also enables automatic code review for all human-created PRs.

## Features

### 1. Automatic CI Failure Detection

The system monitors GitHub webhook events for:
- `check_run` events (individual check completions)
- `check_suite` events (suite of checks completion)

When a linter check fails, the system:
1. Records the failure in the database
2. Extracts failure details from the check run
3. Posts a comment on the PR tagging @copilot with the failure details
4. Tracks the number of fix attempts

### 2. Retry Limitation

To prevent infinite loops, the system:
- Tracks fix attempts per check per PR
- Limits attempts to a maximum of 3 per check
- Stops attempting fixes once the limit is reached

### 3. Automatic Code Review

For all human-created PRs:
- Automatically requests a review from @copilot
- Posts a comment notifying @copilot to review the PR

## Configuration

### Webhook Setup

To enable this feature, configure a GitHub organization webhook with the following settings:

1. **Payload URL**: `https://your-domain.com/api/webhook`
2. **Content type**: `application/json`
3. **Events to subscribe**:
   - Pull requests
   - Check runs
   - Check suites

### Database

The system automatically creates a `pr_check` table to track CI check status:

```sql
CREATE TABLE pr_check (
  id INT AUTO_INCREMENT PRIMARY KEY,
  org VARCHAR(100),
  repo VARCHAR(100),
  pr_number INT,
  check_run_id BIGINT,
  check_name VARCHAR(200),
  status VARCHAR(50),
  conclusion VARCHAR(50),
  failure_reason TEXT,
  fix_attempts INT,
  last_attempt_at DATETIME,
  is_fixed BOOLEAN,
  created_at DATETIME
);
```

### Supported Linter Checks

The system recognizes the following linter check patterns:
- `lint`, `linter`
- `eslint`, `prettier`
- `golangci`, `golint`
- `rubocop`, `pylint`, `flake8`
- `clippy`, `checkstyle`, `pmd`, `spotbugs`
- `style`, `format`

### Configuration Constants

The following constants can be modified in `util/check_api.go`:

- `MaxCheckFailureTextLength`: Maximum length of failure text in comments (default: 500)
- `CopilotUsername`: GitHub username of the copilot bot (default: "copilot")
- `MaxFixAttempts`: Maximum number of fix attempts per check (default: 3)

## API Endpoints

### Get PR Checks

Retrieve all check records for a specific PR:

```
GET /api/get-pr-checks?org=<org>&repo=<repo>&prNumber=<number>
```

**Response:**
```json
[
  {
    "id": 1,
    "org": "casbin",
    "repo": "casbin",
    "prNumber": 1691,
    "checkRunId": 12345678,
    "checkName": "golangci-lint",
    "status": "completed",
    "conclusion": "failure",
    "failureReason": "Check: golangci-lint\nStatus: completed\n...",
    "fixAttempts": 1,
    "lastAttemptAt": "2026-01-25T03:20:00Z",
    "isFixed": false,
    "createdAt": "2026-01-25T03:15:00Z"
  }
]
```

### Get Specific PR Check

Retrieve a specific check record:

```
GET /api/get-pr-check?org=<org>&repo=<repo>&prNumber=<number>&checkName=<name>
```

## Usage Example

### Scenario 1: Linter Failure

1. A developer creates a PR with code that fails the linter check
2. GitHub sends a `check_run` webhook event with `conclusion: failure`
3. The system:
   - Records the failure in the database
   - Posts a comment: "@copilot The CI check has failed. Please help fix the following issue: **Attempt**: 1/3 ..."
4. Copilot receives the notification and can work on fixing the issue
5. If the fix doesn't work and CI fails again, the system repeats steps 3-4
6. After 3 failed attempts, the system stops attempting fixes for that specific check

### Scenario 2: Automatic Code Review

1. A developer creates a new PR
2. GitHub sends a `pull_request` webhook event with `action: opened`
3. The system:
   - Posts a comment: "@copilot Please review this PR."
   - Attempts to request copilot as a reviewer

## Implementation Details

### Key Components

1. **Database Model**: `object/pr_check.go`
   - Manages PR check records
   - Tracks fix attempts and status

2. **GitHub API Utilities**: `util/check_api.go`
   - Retrieves check runs and details
   - Identifies linter checks
   - Posts comments with copilot tags

3. **Webhook Handler**: `controllers/webhook.go`
   - Processes check_run and check_suite events
   - Triggers fix attempts and copilot tagging
   - Enables automatic code review

4. **API Endpoints**: `controllers/pr_check.go`
   - Provides access to PR check records
   - Enables monitoring and debugging

### Error Handling

- Failed API calls are logged but don't crash the system
- Database errors are handled with panic recovery
- Invalid webhook payloads are gracefully ignored

## Limitations

1. The copilot username can be configured via the `CopilotUsername` constant in `util/check_api.go`
2. Only linter checks are monitored (build/test failures are ignored)
3. The system requires proper GitHub webhook configuration
4. GitHub API rate limits may affect functionality in high-traffic scenarios
5. Maximum fix attempts can be configured via the `MaxFixAttempts` constant

## Future Enhancements

Potential improvements:
- Load configuration constants from configuration file instead of code
- Support for custom linter check patterns via configuration
- Dashboard UI for monitoring PR checks
- Notifications for exceeded retry limits
- Integration with Slack/Discord for notifications
- Support for multiple copilot bots
- Webhook signature verification for security
