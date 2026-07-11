#!/usr/bin/env python3
"""
Kaneo API CLI — wraps curl calls for all Kaneo REST endpoints.

Usage:
  kaneo-api.sh projects list --workspace-id <id>
  kaneo-api.sh projects create --workspace-id <id> --name "..." --icon "..." --slug "..."
  kaneo-api.sh projects get --id <id>
  kaneo-api.sh projects update --id <id> --name "..."
  kaneo-api.sh projects delete --id <id>
  kaneo-api.sh projects archive --id <id>
  kaneo-api.sh projects unarchive --id <id>
  kaneo-api.sh tasks list --project-id <id> [--status ...] [--priority ...] [--limit 50]
  kaneo-api.sh tasks create --project-id <id> --title "..." [--priority low] [--status to-do]
  kaneo-api.sh tasks get --id <id>
  kaneo-api.sh tasks update --id <id> [--title "..."] [--status "..."] [--priority "..."]
  kaneo-api.sh tasks delete --id <id>
  kaneo-api.sh tasks move --id <id> --destination-project-id <id> [--destination-status "..."]
  kaneo-api.sh tasks bulk --project-id <id> --task-ids "id1,id2" --operation updateStatus --value "..."
  kaneo-api.sh tasks export --project-id <id>
  kaneo-api.sh tasks import --project-id <id> --file tasks.json

Environment:
  KANEO_API_URL   Base URL (e.g. http://localhost:3000/api)
  KANEO_API_KEY   API key for Bearer auth
"""

import argparse
import json
import os
import sys
import urllib.error
import urllib.request


def api_call(method, path, query=None, body=None):
    """Make an API call and return parsed JSON."""
    base = os.environ.get("KANEO_API_URL", "").rstrip("/")
    api_key = os.environ.get("KANEO_API_KEY", "")

    if not base:
        print("Error: KANEO_API_URL environment variable is not set", file=sys.stderr)
        sys.exit(1)
    if not api_key:
        print("Error: KANEO_API_KEY environment variable is not set", file=sys.stderr)
        sys.exit(1)

    url = f"{base}{path}"
    if query:
        qs = "&".join(f"{k}={urllib.request.quote(str(v))}" for k, v in query.items() if v is not None)
        url = f"{url}?{qs}"

    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json",
    }

    data = json.dumps(body).encode("utf-8") if body is not None else None
    req = urllib.request.Request(url, data=data, headers=headers, method=method)

    try:
        with urllib.request.urlopen(req) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        try:
            err = json.loads(e.read().decode("utf-8"))
            print(f"Error ({e.code}): {err.get('message', e.reason)}", file=sys.stderr)
        except Exception:
            print(f"Error ({e.code}): {e.reason}", file=sys.stderr)
        sys.exit(1)
    except urllib.error.URLError as e:
        print(f"Connection error: {e.reason}", file=sys.stderr)
        sys.exit(1)


# --- Projects ---

def cmd_projects_list(args):
    result = api_call("GET", "/project", query={"workspaceId": args.workspace_id, "includeArchived": args.include_archived})
    print(json.dumps(result, indent=2))

def cmd_projects_create(args):
    body = {"name": args.name, "workspaceId": args.workspace_id, "icon": args.icon, "slug": args.slug}
    result = api_call("POST", "/project", body=body)
    print(json.dumps(result, indent=2))

def cmd_projects_get(args):
    result = api_call("GET", f"/project/{args.id}")
    print(json.dumps(result, indent=2))

def cmd_projects_update(args):
    body = {}
    if args.name: body["name"] = args.name
    if args.icon: body["icon"] = args.icon
    if args.slug: body["slug"] = args.slug
    if args.description: body["description"] = args.description
    if args.is_public is not None: body["isPublic"] = args.is_public
    if not body:
        print("Error: No fields to update. Provide at least one of --name, --icon, --slug, --description, --is-public", file=sys.stderr)
        sys.exit(1)
    result = api_call("PUT", f"/project/{args.id}", body=body)
    print(json.dumps(result, indent=2))

def cmd_projects_delete(args):
    result = api_call("DELETE", f"/project/{args.id}")
    print(json.dumps(result, indent=2))

def cmd_projects_archive(args):
    result = api_call("PUT", f"/project/{args.id}/archive")
    print(json.dumps(result, indent=2))

def cmd_projects_unarchive(args):
    result = api_call("PUT", f"/project/{args.id}/unarchive")
    print(json.dumps(result, indent=2))


# --- Tasks ---

def cmd_tasks_list(args):
    query = {"status": args.status, "priority": args.priority, "assigneeId": args.assignee_id,
             "page": args.page, "limit": args.limit, "sortBy": args.sort_by, "sortOrder": args.sort_order,
             "dueBefore": args.due_before, "dueAfter": args.due_after}
    result = api_call("GET", f"/task/{args.project_id}", query=query)
    print(json.dumps(result, indent=2))

def cmd_tasks_create(args):
    body = {"title": args.title, "status": args.status or "to-do", "priority": args.priority or "no-priority"}
    if args.description: body["description"] = args.description
    if args.start_date: body["startDate"] = args.start_date
    if args.due_date: body["dueDate"] = args.due_date
    if args.user_id: body["userId"] = args.user_id
    result = api_call("POST", f"/task/{args.project_id}", body=body)
    print(json.dumps(result, indent=2))

def cmd_tasks_get(args):
    result = api_call("GET", f"/task/{args.id}")
    print(json.dumps(result, indent=2))

def cmd_tasks_update(args):
    body = {}
    if args.title: body["title"] = args.title
    if args.description is not None: body["description"] = args.description
    if args.status: body["status"] = args.status
    if args.priority: body["priority"] = args.priority
    if args.start_date is not None: body["startDate"] = args.start_date
    if args.due_date is not None: body["dueDate"] = args.due_date
    if args.project_id: body["projectId"] = args.project_id
    if args.position is not None: body["position"] = args.position
    if args.user_id is not None: body["userId"] = args.user_id
    if not body:
        print("Error: No fields to update", file=sys.stderr)
        sys.exit(1)
    result = api_call("PUT", f"/task/{args.id}", body=body)
    print(json.dumps(result, indent=2))

def cmd_tasks_delete(args):
    result = api_call("DELETE", f"/task/{args.id}")
    print(json.dumps(result, indent=2))

def cmd_tasks_move(args):
    body = {"destinationProjectId": args.destination_project_id}
    if args.destination_status: body["destinationStatus"] = args.destination_status
    result = api_call("PUT", f"/task/move/{args.id}", body=body)
    print(json.dumps(result, indent=2))

def cmd_tasks_bulk(args):
    body = {
        "taskIds": args.task_ids.split(","),
        "operation": args.operation,
    }
    if args.value is not None: body["value"] = args.value
    result = api_call("PATCH", "/task/bulk", body=body)
    print(json.dumps(result, indent=2))

def cmd_tasks_export(args):
    result = api_call("GET", f"/task/export/{args.project_id}")
    print(json.dumps(result, indent=2))

def cmd_tasks_import(args):
    try:
        with open(args.file) as f:
            tasks = json.load(f)
    except (FileNotFoundError, json.JSONDecodeError) as e:
        print(f"Error reading file: {e}", file=sys.stderr)
        sys.exit(1)
    body = {"tasks": tasks}
    result = api_call("POST", f"/task/import/{args.project_id}", body=body)
    print(json.dumps(result, indent=2))


# --- Partial updates ---

def cmd_tasks_update_status(args):
    result = api_call("PUT", f"/task/status/{args.id}", body={"status": args.status})
    print(json.dumps(result, indent=2))

def cmd_tasks_update_priority(args):
    result = api_call("PUT", f"/task/priority/{args.id}", body={"priority": args.priority})
    print(json.dumps(result, indent=2))

def cmd_tasks_update_assignee(args):
    result = api_call("PUT", f"/task/assignee/{args.id}", body={"userId": args.user_id})
    print(json.dumps(result, indent=2))

def cmd_tasks_update_due_date(args):
    body = {}
    if args.due_date is not None: body["dueDate"] = args.due_date
    result = api_call("PUT", f"/task/due-date/{args.id}", body=body)
    print(json.dumps(result, indent=2))

def cmd_tasks_update_title(args):
    result = api_call("PUT", f"/task/title/{args.id}", body={"title": args.title})
    print(json.dumps(result, indent=2))

def cmd_tasks_update_description(args):
    result = api_call("PUT", f"/task/description/{args.id}", body={"description": args.description})
    print(json.dumps(result, indent=2))


# --- Parser setup ---

def make_parser():
    parser = argparse.ArgumentParser(description="Kaneo API CLI")
    sub = parser.add_subparsers(dest="command", required=True)

    # projects
    p = sub.add_parser("projects", help="Project operations")
    p_sub = p.add_subparsers(dest="subcommand", required=True)

    p_list = p_sub.add_parser("list", help="List projects")
    p_list.add_argument("--workspace-id", required=True)
    p_list.add_argument("--include-archived", action="store_true")
    p_list.set_defaults(func=cmd_projects_list)

    p_create = p_sub.add_parser("create", help="Create project")
    p_create.add_argument("--workspace-id", required=True)
    p_create.add_argument("--name", required=True)
    p_create.add_argument("--icon", required=True)
    p_create.add_argument("--slug", required=True)
    p_create.set_defaults(func=cmd_projects_create)

    p_get = p_sub.add_parser("get", help="Get project")
    p_get.add_argument("--id", required=True)
    p_get.set_defaults(func=cmd_projects_get)

    p_update = p_sub.add_parser("update", help="Update project")
    p_update.add_argument("--id", required=True)
    p_update.add_argument("--name")
    p_update.add_argument("--icon")
    p_update.add_argument("--slug")
    p_update.add_argument("--description")
    p_update.add_argument("--is-public", action="store_true", default=None)
    p_update.set_defaults(func=cmd_projects_update)

    p_del = p_sub.add_parser("delete", help="Delete project")
    p_del.add_argument("--id", required=True)
    p_del.set_defaults(func=cmd_projects_delete)

    p_arch = p_sub.add_parser("archive", help="Archive project")
    p_arch.add_argument("--id", required=True)
    p_arch.set_defaults(func=cmd_projects_archive)

    p_unarch = p_sub.add_parser("unarchive", help="Unarchive project")
    p_unarch.add_argument("--id", required=True)
    p_unarch.set_defaults(func=cmd_projects_unarchive)

    # tasks
    t = sub.add_parser("tasks", help="Task operations")
    t_sub = t.add_subparsers(dest="subcommand", required=True)

    t_list = t_sub.add_parser("list", help="List tasks")
    t_list.add_argument("--project-id", required=True)
    t_list.add_argument("--status")
    t_list.add_argument("--priority")
    t_list.add_argument("--assignee-id")
    t_list.add_argument("--page", type=int)
    t_list.add_argument("--limit", type=int)
    t_list.add_argument("--sort-by", choices=["createdAt", "priority", "dueDate", "position", "title", "number"])
    t_list.add_argument("--sort-order", choices=["asc", "desc"])
    t_list.add_argument("--due-before")
    t_list.add_argument("--due-after")
    t_list.set_defaults(func=cmd_tasks_list)

    t_create = t_sub.add_parser("create", help="Create task")
    t_create.add_argument("--project-id", required=True)
    t_create.add_argument("--title", required=True)
    t_create.add_argument("--description")
    t_create.add_argument("--status", default="to-do")
    t_create.add_argument("--priority", default="no-priority", choices=["no-priority", "low", "medium", "high", "urgent"])
    t_create.add_argument("--start-date")
    t_create.add_argument("--due-date")
    t_create.add_argument("--user-id")
    t_create.set_defaults(func=cmd_tasks_create)

    t_get = t_sub.add_parser("get", help="Get task")
    t_get.add_argument("--id", required=True)
    t_get.set_defaults(func=cmd_tasks_get)

    t_update = t_sub.add_parser("update", help="Update task (full)")
    t_update.add_argument("--id", required=True)
    t_update.add_argument("--title")
    t_update.add_argument("--description")
    t_update.add_argument("--status")
    t_update.add_argument("--priority", choices=["no-priority", "low", "medium", "high", "urgent"])
    t_update.add_argument("--start-date")
    t_update.add_argument("--due-date")
    t_update.add_argument("--project-id")
    t_update.add_argument("--position", type=int)
    t_update.add_argument("--user-id")
    t_update.set_defaults(func=cmd_tasks_update)

    t_del = t_sub.add_parser("delete", help="Delete task")
    t_del.add_argument("--id", required=True)
    t_del.set_defaults(func=cmd_tasks_delete)

    t_move = t_sub.add_parser("move", help="Move task to another project")
    t_move.add_argument("--id", required=True)
    t_move.add_argument("--destination-project-id", required=True)
    t_move.add_argument("--destination-status")
    t_move.set_defaults(func=cmd_tasks_move)

    t_bulk = t_sub.add_parser("bulk", help="Bulk update tasks")
    t_bulk.add_argument("--task-ids", required=True, help="Comma-separated task IDs")
    t_bulk.add_argument("--operation", required=True,
                        choices=["updateStatus", "updatePriority", "updateAssignee", "delete", "addLabel", "removeLabel", "updateDueDate"])
    t_bulk.add_argument("--value")
    t_bulk.set_defaults(func=cmd_tasks_bulk)

    t_export = t_sub.add_parser("export", help="Export tasks")
    t_export.add_argument("--project-id", required=True)
    t_export.set_defaults(func=cmd_tasks_export)

    t_import = t_sub.add_parser("import", help="Import tasks from JSON file")
    t_import.add_argument("--project-id", required=True)
    t_import.add_argument("--file", required=True, help="Path to JSON file with tasks array")
    t_import.set_defaults(func=cmd_tasks_import)

    # Partial updates (hidden from help but available)
    t_status = t_sub.add_parser("update-status", help="Update task status only")
    t_status.add_argument("--id", required=True)
    t_status.add_argument("--status", required=True)
    t_status.set_defaults(func=cmd_tasks_update_status)

    t_pri = t_sub.add_parser("update-priority", help="Update task priority only")
    t_pri.add_argument("--id", required=True)
    t_pri.add_argument("--priority", required=True, choices=["no-priority", "low", "medium", "high", "urgent"])
    t_pri.set_defaults(func=cmd_tasks_update_priority)

    t_assign = t_sub.add_parser("update-assignee", help="Update task assignee only")
    t_assign.add_argument("--id", required=True)
    t_assign.add_argument("--user-id", required=True)
    t_assign.set_defaults(func=cmd_tasks_update_assignee)

    t_due = t_sub.add_parser("update-due-date", help="Update task due date only")
    t_due.add_argument("--id", required=True)
    t_due.add_argument("--due-date")
    t_due.set_defaults(func=cmd_tasks_update_due_date)

    t_title = t_sub.add_parser("update-title", help="Update task title only")
    t_title.add_argument("--id", required=True)
    t_title.add_argument("--title", required=True)
    t_title.set_defaults(func=cmd_tasks_update_title)

    t_desc = t_sub.add_parser("update-description", help="Update task description only")
    t_desc.add_argument("--id", required=True)
    t_desc.add_argument("--description", required=True)
    t_desc.set_defaults(func=cmd_tasks_update_description)

    return parser


def main():
    parser = make_parser()
    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
