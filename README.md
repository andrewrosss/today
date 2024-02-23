# today

A small script (`today`) to create a rolling todo list.

The command will copy the yesterday's todo list file into today's todo list file, letting you pick up fom where you left off from the last todo-list you had.

Taken from this [hackernews comment](https://news.ycombinator.com/item?id=39433880). From the HN comment:

> ... `today` basically opens `${TODO_HOME}/$YYYY_MM_DD-todo.txt`, but it'll start you off with a copy of the most recent (previous) file.
>
> This lets me have "durable" files (I can grep for pretty much anything and get a date-specific hit for it, similar to doing a `git log -S`), and also lets me declare "task-bankruptcy" without any worry (I can always "rewind" to any particular point in time).
