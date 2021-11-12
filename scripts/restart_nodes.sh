docker-compose -f "../docker-compose.yaml" restart node1 node2 node3 node4 node5 node6 node7 && docker-compose -f "../docker-compose.yaml" logs -f -t | grep -v ganache
