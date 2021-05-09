docker-compose up -d --build --scale node=7 --force-recreate
#docker restart $(docker ps -a --format "{{.Names}}") # | grep node | sort -n)
docker-compose logs -f -t | grep -v ganache
