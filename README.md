## децентрализованный релеер для выполнения кросc-чейн вызовов


###  локальный деплой

```bash
mkdir digiu
cd digiu
git clone git@github.com:DigiU-Lab/p2p-bridge.git
git clone git@github.com:DigiU-Lab/eth-contracts.git
cd p2p-bridge
./deploy.sh
```


### запуск теста
 
```
 cd ../eth-contracts/truffle
 npm run integration-test:local
 
```

