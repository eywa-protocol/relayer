nodes=$(docker ps -a --format "{{.Names}}" | grep node | sort -n)
if [[ "$nodes" == "" ]] ;then
  echo "nodes containers not found"
else
  docker stop $nodes
  docker rm $nodes
fi

gsns=$(docker ps -a --format "{{.Names}}" | grep gsn | sort -n)
if [[ "$gsns" == "" ]] ;then
  echo "gsn nodes containers not found"
else
  docker stop $gsns
  docker rm $gsns
fi


bsns=$(docker ps -a --format "{{.Names}}" | grep bsn | sort -n)
if [[ "$bsns" == "" ]] ;then
  echo "bootstrap nodes containers not found"
else
  docker stop $bsns
  docker rm $bsns
fi


if [[ "$(docker images p2p-bridge_img | grep p2p-bridge_img)" == "" ]];then
  echo "nodes image not found"
else
  docker rmi -f p2p-bridge_img
fi

make -C ./.. clean
