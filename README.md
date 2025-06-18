# git-aware-todo-cli

A minimal terminal-based todo app that stores tasks in a JSON file and optionally runs Git commands when tasks are completed.

---

## Features

- CLI interface with commands: `add`, `ls`, `done`, `rm`, `init`
- Tasks stored locally in `.todo.json`
- Mark tasks as important
- Run Git or shell commands automatically when a task is marked done
- Filter tasks by completion status or importance
- Simple table view for task listing

---

## Usage

```bash
# Initialize the task file
todo init

# Add a new task
todo add

# List tasks (all, completed, uncompleted, important)
todo ls
todo ls --com
todo ls --uncom
todo ls --imp

# Mark a task as completed
todo done <task_id>

# Delete the todo list
todo rm
```

