#!/bin/bash
# ForceCommand handler — attach to existing tmux dev session or create one
SESSION="dev"

if tmux has-session -t "$SESSION" 2>/dev/null; then
    exec tmux attach -t "$SESSION"
fi

# Create the standard dev layout (same as start-session.sh but with fixed dimensions)
tmux new-session -d -s "$SESSION" -x 200 -y 50 nvim
tmux split-window -h -t "$SESSION" -p 35 /bin/bash
tmux split-window -v -t "$SESSION" claude
tmux select-pane -t "$SESSION:0.0"
exec tmux attach -t "$SESSION"
