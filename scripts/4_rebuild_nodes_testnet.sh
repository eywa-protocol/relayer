make -C ../ clean
export MODE=init

export ECDSA_KEY_1=0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268,0xf5e350eb75d845a8cd538e06331ee6eeb159c54710c6c84c725cd77e2e0dde59,0x9894defbe159c1abba7db3f88b122bf94c6838a28a98af9466e85c8f573c43bc,0xafd8124dd3abec91d07dca54878cd296666a96decd122cc7646e9357df21a6de,0xb433482fbde7e7a9f6fb0d63d60782820dfe8e1f2e58ac9c0281c817d13b64ec,0xa76afd42e10c31cc170f32489a9f6af28efe2b1e6aede2813c3ec7d2907a05e9,0x7eb8b227320968eefa7291cfb5ff81226301b8fb40dc2eace533ec594ddb79e1
export ECDSA_KEY_2=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7,0xe84538a8b4059da88d5a76544329093da99603fe05b0f1332f737c74253688c3,0x469e5c05e289274dd8570c31f2f0f21236f2e071613ac9c565821985e7ae641e,0x95472b385de2c871fb293f07e76a56e8e93ea4e743fe940afbd44c30730211dc,0xd37fe8ab9b9a4f0473c2c7e32b6b218cc398837ce29a548066d52245c345df1a,0xdd9a03e0e395072e4f8b96952264a3cb30efa28bf98342e197d2af8a3d6fc541,0xf8d3535d747014e91637e8b9dcb88f265d13b6219c2c2a93df94659303c42d64
export ECDSA_KEY_3=0x2211abbef7d05ba31bfe06286d4ca036a72a0c1ebf2ffcc1f95ad8d8e53b089f

export NETWORK_RPC_1=wss://rinkeby.infura.io/ws/v3/ab95bf9f6dd743e6a8526579b76fe358
export NETWORK_RPC_2=ws://95.217.104.54:8576
export NETWORK_RPC_3=wss://rpc-mumbai.maticvigil.com/ws

#export RANDEVOUE=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1)
export RANDEVOUE="jQrKHTb7whOqy7oOZmCGbpoHHcSJDgs7DqXRonn35H2gMnFSsfDN3qc"


docker-compose -f ../docker-compose-testnet.yaml up -d --build --no-deps
#docker-compose -f ../docker-compose-testnet.yaml up -d

#docker restart $(docker ps -a --format "{{.Names}}" | grep node | sort -n)
#  docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
docker-compose -f ../docker-compose-testnet.yaml logs -f -t | grep -v ganache

