# Gira CLI

[![JSR](https://jsr.io/badges/@mklinovsky/gira)](https://jsr.io/@mklinovsky/gira)

Gira is a command-line tool designed to streamline the management of GitLab and
JIRA tasks. It allows you to create JIRA issues, change their status, and create
merge requests with ease.

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

### With Deno

Ensure you have [Deno](https://deno.com/) installed on your system.

Then, install Gira using the `deno install` command, specifying permissions
individually for more fine-grained control:

```bash
deno install -g --allow-env --allow-sys --allow-read --allow-net --allow-run -n gira jsr:@mklinovsky/gira
```

- `-g`: Installs the executable globally.
- `--allow-net`: Allows network access for API calls to Jira and GitLab.
- `--allow-env`: Allows access to environment variables for API tokens and URLs.
- `--allow-sys`: Allows access to system information.
- `--allow-read`: Allows reading files.
- `--allow-run`: Allows running subprocesses (e.g., `pbcopy` for copying to
  clipboard).
- `-n gira`: Specifies the executable name as `gira`.

Alternatively, you can grant all permissions (less recommended for security
reasons):

```bash
deno install -g -A -n gira jsr:@mklinovsky/gira
```

- `-A`: Grants all permissions (network, environment variables, etc.) required
  by the CLI. This is a convenient alternative to specifying individual
  permissions.

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

Create a merge request for the current branch targeting the master branch. It
also updates the JIRA issue status to "In Review".

```bash
gira mr [options]
```

**Options:**

- `-t, --target <target>`: Specify the target branch for the merge request
  (default: master).
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
