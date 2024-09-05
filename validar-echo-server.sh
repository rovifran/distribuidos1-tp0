#!/bin/bash
expected_response="Test message"
response=$(docker run -i --name=echo_server_test --network=tp0_testing_net alpine sh -c "echo $expected_response | nc -w 10 server 12345")

if [ "$expected_response" = "$response" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

docker stop echo_server_test
docker rm echo_server_test