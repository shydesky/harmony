LEADER_PORT=3333
SLAVE_START_PORT=4000
SLAVE_END_PORT=4009
for port in `seq $SLAVE_START_PORT $SLAVE_END_PORT`;
do
    go run ./benchmark_node.go -mode slave -port $port &
done
go run ./benchmark_node.go -mode leader -port $LEADER_PORT &