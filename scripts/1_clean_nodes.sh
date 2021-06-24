nodes=$(docker ps -a --format "{{.Names}}" | grep node | sort -n)
if [[ "$nodes" == "" ]] ;then
  echo "nodes containers not found"
else
  docker stop $nodes
  docker rm $nodes
fi
#docker rmi -f $nodes
if [[ "$(docker images p2p-bridge_node | grep p2p-bridge_node)" == "" ]];then
  echo "nodes images not found"
else
  docker rmi -f p2p-bridge_node
fi

bsns=$(docker ps -a --format "{{.Names}}" | grep bsn | sort -n)
if [[ "$bsns" == "" ]] ;then
  echo "bootstrap nodes containers not found"
else
  docker stop $bsns
  docker rm $bsns
fi
#docker rmi -f $nodes
if [[ "$(docker images p2p-bridge_bsn | grep p2p-bridge_bsn)" == "" ]];then
  echo "bootstrap nodes image not found"
else
  docker rmi -f p2p-bridge_bsn
fi