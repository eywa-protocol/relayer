nodes=$(docker ps -a --format "{{.Names}}" | grep node | sort -n)
if [[ "$nodes" == "" ]] ;then
  echo "nodes containers not found"
else
  docker stop $nodes
  docker rm $nodes
fi

bsns=$(docker ps -a --format "{{.Names}}" | grep bsn | sort -n)
if [[ "$bsns" == "" ]] ;then
  echo "bootstrap nodes containers not found"
else
  docker stop $bsns
  docker rm $bsns
fi


if [[ "$(docker images p2p-bridge_img | grep p2p-bridge_img)" == "" ]];then
  echo "bootstrap nodes image not found"
else
  docker rmi -f p2p-bridge_img
fi
rm -rf ../.data/*