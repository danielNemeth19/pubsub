#! /bin/fish
set sessions "code" "rabbit" "server" "p1" "p2"

for i in $sessions
    tns $i
    tmux send-keys -t $i:0 'cd /go/src/pubsub' C-m
end

tas $sessions[1]
