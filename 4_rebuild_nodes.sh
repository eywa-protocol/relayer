make clean
export MODE=init
docker-compose up -d --build --no-deps node
docker-compose up -d --scale node=17
#docker restart $(docker ps -a --format "{{.Names}}" | grep node | sort -n)
docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
docker-compose logs -f -t | grep -v ganache

# # FOR SINGLE MODE
# make clean
# export MODE=singlenode
# docker-compose up -d --build


