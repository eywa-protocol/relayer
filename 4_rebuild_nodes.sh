make clean
docker-compose up -d --build --no-deps node
docker-compose up -d --scale node=7
docker restart $(docker ps -a --format "{{.Names}}" | grep node | sort -n)
docker-compose logs -f -t | grep -v ganache


