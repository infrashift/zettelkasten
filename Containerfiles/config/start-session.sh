#!/bin/bash
SESSION="dev"

tmux new-session -d -s "$SESSION" -x "$(tput cols)" -y "$(tput lines)" nvim

# Split right pane (35% width)
tmux split-window -h -t "$SESSION" -p 35 /bin/bash

# Split that right pane horizontally for claude
tmux split-window -v -t "$SESSION" claude

# Focus left pane (nvim)
tmux select-pane -t "$SESSION:0.0"

tmux attach -t "$SESSION"
