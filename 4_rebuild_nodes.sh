make clean
docker-compose up -d --build --no-deps node
docker-compose up -d --scale node=30
#docker restart $(docker ps -a --format "{{.Names}}" | grep node | sort -n)
docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
docker-compose logs -f -t | grep -v ganache


