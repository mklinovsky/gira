# Gira CLI

Gira is a command-line tool designed to streamline the management of GitLab and
JIRA tasks. It allows you to create JIRA issues, change their status, and create
merge requests with ease.

## Installation

To use Gira, ensure you have [Deno](https://deno.com/) installed on your system.

Then, install Gira using the `deno install` command:

```bash
deno install -g -A -n gira jsr:@mklinovsky/gira
```

- `-g`: Installs the executable globally.
- `-A`: Grants all permissions (network, environment variables, etc.) required
  by the CLI. This is a convenient alternative to specifying individual
  permissions.
- `-n gira`: Specifies the name of the executable as `gira`.

Alternatively, you can specify permissions individually for more fine-grained
control:

```bash
deno install -g --allow-net --allow-env --allow-run -n gira jsr:@mklinovsky/gira
```

- `--allow-net`: Allows network access for API calls to Jira and GitLab.
- `--allow-env`: Allows access to environment variables for API tokens and URLs.
- `--allow-run`: Allows running subprocesses (e.g., `pbcopy` for copying to
  clipboard).
- `-n gira`: Specifies the executable name as `gira`.

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
- `-t, --type <type>`: Specify the issue type (default: Task).
- `-b, --branch`: Create a corresponding Git branch.
- `-a, --assign`: Assign the issue to yourself.
- `-s, --start`: Start progress on the issue.

### Change JIRA Issue Status

Change the status of an existing JIRA issue. If no issue key is provided, it
attempts to derive it from the current Git branch name.

```bash
gira status <status> [options]
```

**Options:**

- `-i, --issue <issue>`: Specify the issue key.

### Create a Merge Request

Create a merge request for the current branch targeting the master branch. It
also updates the JIRA issue status to "In Review".

```bash
gira mr
```

### Usual Workflow

```bash
gira create -b -a -s "Fix it"
```

Will create a new JIRA issue with the summary "Fix it", create a Git branch with
the name prefixed by the JIRA issue key (proj-123-fix-it), assign the issue to
yourself, and start progress on it.

```bash
gira mr
```

Will create a merge request for the current branch, targeting the master branch,
and update the JIRA issue status to "In Review".

```bash
gira status Done
```

Will change the status of the JIRA issue to "Done". If no issue key is provided,
it will attempt to derive it from the current Git branch name.

## Environment Variables

To use Gira, you need to set the following environment variables:

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
