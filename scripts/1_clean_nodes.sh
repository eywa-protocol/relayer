nodes=$(docker ps -a --format "{{.Names}}" | grep node | sort -n)
docker stop $nodes
docker rm $nodes
docker rmi -f $nodes
