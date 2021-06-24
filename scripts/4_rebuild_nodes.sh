#!/usr/bin/env bash

make -C ../ clean
export MODE=init
export ECDSA_KEY_1=0x60cc6f7a3b09e5080dc86cc0fd80e29545683ad4336012b221998b448d2d57bb,0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268,0xf5e350eb75d845a8cd538e06331ee6eeb159c54710c6c84c725cd77e2e0dde59,0x9894defbe159c1abba7db3f88b122bf94c6838a28a98af9466e85c8f573c43bc,0xafd8124dd3abec91d07dca54878cd296666a96decd122cc7646e9357df21a6de,0xb433482fbde7e7a9f6fb0d63d60782820dfe8e1f2e58ac9c0281c817d13b64ec,0xa76afd42e10c31cc170f32489a9f6af28efe2b1e6aede2813c3ec7d2907a05e9,0x7eb8b227320968eefa7291cfb5ff81226301b8fb40dc2eace533ec594ddb79e1,0x594e53edfffd4854087d036a8b5a768904be914855ea7985ad094da2f59ce17a,0xa5f9ca32c5fcaf7625a7207261ef274697a2b6ec419050b7165546b99288bc07,0x6926955d28a178b084a54cebd19e7d0107fe02cd9fd5d4f37d430909d90b0259,0x69dbe83e3d9fd2ccef9f10c6cde17b7389e14b68d2b0c6d49dc5a0632ca011ce,0x39f19403a5fb8da5a2afdaad4ed4fe15871d9691e3e8efebdbb715d31736535b,0x611065277d9fac2ebe688d0797e07bc01bd1d0121f630b1ddf7bd5f356d4416e,0xad653f2e32184fe2f5e79c8c745cbca0be4205e3bb48242c75542210a29b7fcb
export ECDSA_KEY_2=0x72a21bf9881eff3ab6e4bd245074cdbb6631388e22318bd4f5df2752b69222cc,0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7,0xe84538a8b4059da88d5a76544329093da99603fe05b0f1332f737c74253688c3,0x469e5c05e289274dd8570c31f2f0f21236f2e071613ac9c565821985e7ae641e,0x95472b385de2c871fb293f07e76a56e8e93ea4e743fe940afbd44c30730211dc,0xd37fe8ab9b9a4f0473c2c7e32b6b218cc398837ce29a548066d52245c345df1a,0xdd9a03e0e395072e4f8b96952264a3cb30efa28bf98342e197d2af8a3d6fc541,0xf8d3535d747014e91637e8b9dcb88f265d13b6219c2c2a93df94659303c42d64,0xef0b0807515081cb8a5c171a1944bc07c1dd27d2184590ac150a7713d335f47a,0x01b039c047c67e23305400d0fec71a79aa0130a8f15dc4d4698281bf6612df2f,0x13e8318ea3359711e7106712f556ec99be0fff058f5017197f3fb5dbadae1f16,0xcc90be77110f973d3a2ccc99b4657e1976e1d3eeb0472a774dc8b6eb981ebb2e,0x07f386df7ea062283e5fa37f52da002aa5eff5a214ba60a1365cbad4694e0843,0x60a82fb2c0ef40d68c81b522b3a74f5ccb742aaf076c8a99b1b1f1d941ce937c,0x63fafc92fd796cbc92fb4d0ba0fdd8db4ae832d74b30c743487df5f95fc793ce
export ECDSA_KEY_3=0x72a21bf9881eff3ab6e4bd245074cdbb6631388e22318bd4f5df2752b69222cc,0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7,0xe84538a8b4059da88d5a76544329093da99603fe05b0f1332f737c74253688c3,0x469e5c05e289274dd8570c31f2f0f21236f2e071613ac9c565821985e7ae641e,0x95472b385de2c871fb293f07e76a56e8e93ea4e743fe940afbd44c30730211dc,0xd37fe8ab9b9a4f0473c2c7e32b6b218cc398837ce29a548066d52245c345df1a,0xdd9a03e0e395072e4f8b96952264a3cb30efa28bf98342e197d2af8a3d6fc541,0xf8d3535d747014e91637e8b9dcb88f265d13b6219c2c2a93df94659303c42d64,0xef0b0807515081cb8a5c171a1944bc07c1dd27d2184590ac150a7713d335f47a,0x01b039c047c67e23305400d0fec71a79aa0130a8f15dc4d4698281bf6612df2f,0x13e8318ea3359711e7106712f556ec99be0fff058f5017197f3fb5dbadae1f16,0xcc90be77110f973d3a2ccc99b4657e1976e1d3eeb0472a774dc8b6eb981ebb2e,0x07f386df7ea062283e5fa37f52da002aa5eff5a214ba60a1365cbad4694e0843,0x60a82fb2c0ef40d68c81b522b3a74f5ccb742aaf076c8a99b1b1f1d941ce937c,0x63fafc92fd796cbc92fb4d0ba0fdd8db4ae832d74b30c743487df5f95fc793ce

export NETWORK_RPC_1=ws://172.20.128.11:7545
export NETWORK_RPC_2=ws://172.20.128.12:8545
export NETWORK_RPC_3=ws://172.20.128.13:9545

if [[ "$OSTYPE" == "darwin"* ]];then
  export RANDEVOUE=$(LC_CTYPE=C tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w ${1:-32} | head -n 1)
  docker-compose -f ../docker-compose-macos.yaml up -d --build --no-deps node
  docker-compose -f ../docker-compose-macos.yaml up -d --scale node=7
  #docker restart $(docker ps -a --format "{{.Names}}" | grep node | sort -n)
  #docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
  docker-compose -f ../docker-compose-macos.yaml logs -f -t | grep -v ganache
else
  export RANDEVOUE=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1)
  docker-compose -f ../docker-compose.yaml up -d --build --no-deps node
  docker-compose -f ../docker-compose.yaml up -d --scale node=7
  #docker restart $(docker ps -a --format "{{.Names}}" | grep node | sort -n)
  #docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
  docker-compose -f ../docker-compose.yaml logs -f -t | grep -v ganache
fi

# # FOR SINGLE MODE
# make clean
# export MODE=singlenode
# export ECDSA_KEY_1=0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268
# export ECDSA_KEY_2=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7
# export NETWORK_RPC_1=ws://172.20.128.11:7545
# export NETWORK_RPC_2=ws://172.20.128.12:8545
# docker-compose up -d --build node
