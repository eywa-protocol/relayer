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
$ export TYPE_ADAPTER_ENV=bootstrap.env && export ECDSA_KEY_1=0x60cc6f7a3b09e5080dc86cc0fd80e29545683ad4336012b221998b448d2d57bb && export ECDSA_KEY_2=0x469e5c05e289274dd8570c31f2f0f21236f2e071613ac9c565821985e7ae641e  && servicename=dev-adapter make develop
```

### Пример запуска теста в режиме single mode

```json
$ echo "в файле 4_rebuild_nodes.sh раскоментарить нужный кусок скрипта"
$ cd p2p-bridge
$ ./deploy.sh
$ cd eth-contracts/truffle
$ npm run integration-test:local
```
  



