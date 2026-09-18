# Gira CLI

Gira is a command-line tool designed to streamline the management of GitLab and
JIRA tasks. It allows you to create JIRA issues, change their status, and create
merge requests with ease.

It can also load per-project settings from `~/.gira/config.json`, so one install
can work across Jira and non-Jira repositories.

## Installation

### Using install script

You can install `gira` by running the following command in your terminal. The
script will download and install the correct binary for your system into
`~/.gira/bin`.

```bash
curl -fsSL https://raw.githubusercontent.com/mklinovsky/gira/main/scripts/install.sh | bash
```

The script will also attempt to add the installation directory to your shell's
`PATH`. If it cannot, it will provide instructions on how to do so manually.

### With Go

Ensure you have [Go](https://go.dev/) 1.25 or newer installed, then:

```bash
go install github.com/mklinovsky/gira@latest
```

### From source

```bash
git clone https://github.com/mklinovsky/gira.git
cd gira
go build -o gira .
```

## Usage

The CLI provides several commands to interact with JIRA and GitLab:

### Create a JIRA Issue

Create a new JIRA issue with an optional parent issue, type, and additional
actions like branch creation and assignment.

```bash
gira create <summary> [options]
```

**Options:**

- `-p, --parent <parent>`: Specify a parent issue key.
- `-t, --type <type>`: Specify the issue type. If omitted, gira uses configured
  defaults and parent hierarchy.
- `-b, --branch`: Create a corresponding Git branch.
- `-w, --worktree <directory>`: Create a Git worktree in the specified base
  directory (mutually exclusive with `-b`).
- `-a, --assign`: Assign the issue to yourself.
- `-s, --start`: Start progress on the issue.

### Get a JIRA Issue

Get a JIRA issue by its key.

```bash
gira get-jira <issue-key>
```

### Get JIRA Issue Attachments

Download attachments from a JIRA issue.

```bash
gira get-jira-files <issue-key> [options]
```

**Options:**

- `-o, --output-dir <output-dir>`: The directory to download the files to.

### Update a JIRA Issue

Update a JIRA issue with a custom field.

```bash
gira update-jira <issue-key> [options]
```

**Options:**

- `-c, --custom-field <custom-field>`: A custom field to update, in the format
  `key=value`.

### Change JIRA Issue Status

Change the status of an existing JIRA issue. If no issue key is provided, it
attempts to derive it from the current Git branch name.

```bash
gira status <status> [options]
```

**Options:**

- `-i, --issue <issue>`: Specify the issue key.

### Create a Merge Request

Create a merge request for current branch. Target branch defaults to
`defaults.gitlab.targetBranch` or matched project `gitlab.targetBranch` when
configured, else falls back to `master`. It also updates JIRA issue status to
"In Review" when Jira is enabled for current folder and branch name starts with
Jira key.

```bash
gira mr [options]
```

**Options:**

- `-t, --target <target>`: Specify target branch for merge request. Overrides
  configured `gitlab.targetBranch`.
- `--title <title>`: Explicit merge request title. If omitted, Gira derives the
  title from branch name.
- `-l, --labels <labels>`: Comma-separated labels for the merge request.
- `-d, --draft`: Create a draft merge request.

### Get a Merge Request

Get a merge request by its ID.

```bash
gira get-mr <merge-request-id>
```

### Merge a Merge Request

Merge a merge request by its ID.

```bash
gira merge <merge-request-id> [options]
```

**Options:**

- `--close-jira`: Close the associated JIRA issue.
- `--delete-branch`: Delete the branch after merging.

## Config File

Gira reads config from `~/.gira/config.json`.

Create it with:

```bash
gira init
```

Overwrite existing config with:

```bash
gira init --force
```

Template:

```json
{
  "defaults": {
    "gitlab": {
      "url": "",
      "apiToken": "",
      "projectId": "",
      "userId": "",
      "targetBranch": ""
    },
    "jira": {
      "enabled": true,
      "url": "",
      "apiToken": "",
      "userEmail": "",
      "userId": "",
      "projectKey": "",
      "issueType": "",
      "subtaskIssueType": ""
    }
  },
  "projects": []
}
```

Empty string values are ignored, so they do not override environment variables.

### Per-project settings

`projects[].path` is matched against current working directory. Longest matching
path wins.

If you keep linked worktrees outside project root, set `worktreeBasePath` on the
project entry so commands like `gira mr` still resolve same project config when
run inside worktree directory.

```json
{
  "defaults": {
    "gitlab": {
      "url": "https://gitlab.example.com",
      "apiToken": "gitlab-token",
      "userId": "123",
      "targetBranch": "main"
    },
    "jira": {
      "enabled": true,
      "url": "https://company.atlassian.net",
      "apiToken": "jira-token",
      "userEmail": "me@example.com",
      "userId": "jira-user-id",
      "issueType": "Task",
      "subtaskIssueType": "Sub-task"
    }
  },
  "projects": [
    {
      "path": "~/Projects/app-one",
      "worktreeBasePath": "~/Projects/app-one-worktrees",
      "gitlab": {
        "projectId": "111",
        "targetBranch": "release"
      },
      "jira": {
        "projectKey": "APP",
        "issueType": "Task",
        "subtaskIssueType": "Sub-task"
      }
    },
    {
      "path": "~/Projects/internal-tool",
      "gitlab": {
        "projectId": "222"
      },
      "jira": {
        "enabled": false
      }
    }
  ]
}
```

When `gira create` runs without `--type`, `jira.issueType` is used for
standalone issues and children of epics, while `jira.subtaskIssueType` is used
for children of standard issues.

Resolution order:

- environment variables
- `defaults`
- matched `projects[]` entry
- CLI flags

Final precedence is `CLI > project > defaults > env`.

When Jira is disabled for current folder:

- `gira mr` still creates merge requests
- `gira merge` still merges merge requests
- Jira-only commands fail with a clear error

### Usual Workflow

```bash
gira create -b -a -s "Fix it"
```

Will create a new JIRA issue with the summary "Fix it", create a Git branch with
the name prefixed by the JIRA issue key (proj-123-fix-it), assign the issue to
yourself, and start progress on it.

```bash
gira create -w ../worktrees -a -s "Fix it"
```

Will create a new JIRA issue with the summary "Fix it", create a Git worktree in
`../worktrees/proj-123-fix-it`, assign the issue to yourself, and start progress
on it.

```bash
gira mr
```

Will create a merge request for the current branch, targeting the master branch,
and update the JIRA issue status to "In Review" when branch name starts with a
Jira key.

```bash
gira mr --title "Release 1.2.0"
```

Will create a merge request with explicit title `Release 1.2.0`.

Without `--title`, non-Jira branch names are converted to readable titles. For
example, `feat/mobile/login-flow` becomes `feat: Mobile login flow`.

```bash
gira status Done
```

Will change the status of the JIRA issue to "Done". If no issue key is provided,
it will attempt to derive it from the current Git branch name.

## Environment Variables

You can still configure Gira with environment variables. They are used as
fallback values when config file does not provide them:

- `JIRA_API_TOKEN`: Your JIRA API token.
- `JIRA_URL`: The base URL for your JIRA instance.
- `JIRA_USER_EMAIL`: Your JIRA user email.
- `JIRA_USER_ID`: Your JIRA user ID.
- `JIRA_PROJECT_KEY`: Your JIRA project key.
- `GITLAB_API_TOKEN`: Your GitLab API token.
- `GITLAB_URL`: The base URL for your GitLab instance.
- `GITLAB_PROJECT_ID`: Your GitLab project ID.
- `GITLAB_USER_ID`: Your GitLab user ID.

## License

This project is licensed under the MIT License.
