# today

A small script (`today`) to create a rolling todo list.

The command will copy the yesterday's todo list file into today's todo list file, letting you pick up fom where you left off from the last todo-list you had.

Taken from this [hackernews comment](https://news.ycombinator.com/item?id=39433880). From the HN comment:

> ... `today` basically opens `${TODO_HOME}/$YYYY_MM_DD-todo.txt`, but it'll start you off with a copy of the most recent (previous) file.
>
> This lets me have "durable" files (I can grep for pretty much anything and get a date-specific hit for it, similar to doing a `git log -S`), and also lets me declare "task-bankruptcy" without any worry (I can always "rewind" to any particular point in time).

## Installation

1. Copy the three files (`today`, `previous`, `report`) somewhere on your machine — make sure they're on your `$PATH`.
1. In all 3 files replace:
   - `SCRIPT_DIR` with the directory that you copied the scripts into.
   - `TODO_HOME` with the directory that you want the todo files to be written.

    For example, these are the settings on my machine:

    ```bash
    SCRIPT_DIR="$HOME/.local/bin"
    TODO_HOME="$HOME/Documents/sandbox/todo"
    ```

## Usage

```bash
$ today  # generates a new todo file for today
/Users/andrewrosss/Documents/sandbox/todo/todo-2024-02-23-todo.md
```

Or open the new file in your favourite code editor:

```bash
$ code -n $(today) # open today's todo file in vscode
```