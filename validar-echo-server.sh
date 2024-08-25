#!/bin/bash
docker run --detach --name=echo-server-test --network=tp0_testing_net subfuzion/netcat -l -p 12345
sleep 2
echo -n "Test message" | docker exec -i echo-server-test nc -N server 12345
