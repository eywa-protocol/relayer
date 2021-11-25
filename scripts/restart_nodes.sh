if [[ "$OSTYPE" == "darwin"* ]];then
  DC="../docker-compose-macos.yaml"
else
  DC="../docker-compose.yaml"
fi

docker-compose -f $DC restart node1 node2 node3 node4 node5 node6 node7
docker-compose -f $DC logs -f -t | grep -v ganache
