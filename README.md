### Пример добавления Node в список через скрипт

```bash
cd ./eth-contracts/truffle
npx truffle exec './scripts/init/addNodeToList.js' --network network1 --wallet_node "0x6786D7A5d7f1898220503aa35527250B275dBBE9" --pubkey_node "0x194e868506502e5ecae3e5b623801548125a748e6b2da15681312a7cf0283acc" --p2p_address "/ip4/127.0.0.1/tcp/6666/p2p/QmbDd2omyRkTipjw6jTjrHYKp2pPhKp15SNZ2azqvvp1i7" --enable "true"
```



### Пример добавления Node в список через js код

```json
this.p2pAddressNode_1 = await web3.utils.fromAscii('/ip4/127.0.0.1/tcp/6666/p2p/QmbDd2omyRkTipjw6jTjrHYKp2pPhKp15SNZ2azqvvp1i7');
this.p2pAddressNode_2 = await web3.utils.fromAscii('/ip4/127.0.0.1/tcp/7777/p2p/DsqDd2omyRkTipjw6jTjrHYKp2pPhKp15SNZ2azqvvp1i8');
this.pubKeyNode_1     = '0x194e868506502e5ecae3e5b623801548125a748e6b2da15681312a7cf0283acc';
this.pubKeyNode_2     = '0x0da632a0af66bc9748f4fe4e8261facffbaef084ae1c591b1d30889622975735';
this.enableNode_1     = true;
this.enableNode_2     = true;

const tx1 = await this.nodeList.addNode(adrWalletNode1, this.p2pAddressNode_1, this.pubKeyNode_1, this.enableNode_1);
```


### Пример работы с air

```json
$ docker-compose stop dev-adapter && docker-compose rm dev-adapter && docker rmi p2p-bridge_dev-adapter -f # optional
$ export TYPE_ADAPTER_ENV=bootstrap.env && export ECDSA_KEY_1=0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268 && export ECDSA_KEY_2=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7 && export NETWORK_RPC_1=wss://rinkeby.infura.io/ws/v3/ab95bf9f6dd743e6a8526579b76fe358 && export NETWORK_RPC_2=ws://95.217.104.54:8576 && servicename=dev-adapter make develop

```

### Пример запуска теста в режиме single mode в ЛОКАЛЬНОЙ СРЕДЕ

```json
$ echo "в файле 4_rebuild_nodes.sh раскоментарить нужный кусок скрипта"
$ cd p2p-bridge
$ ./deploy.sh
$ cd eth-contracts/truffle
$ npm run integration-test:local
```

### Пример запуска теста в режиме single mode в TESTNET

```json
$ cd p2p-bridge
$ ./deploy.testnet.singlemode.sh
$ cd eth-contracts/truffle
$ npm run integration-test:testnet
```
  



