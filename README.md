## децентрализованный релеер для выполнения кросc-чейн вызовов


###  локальный деплой

```bash
mkdir digiu
cd digiu
git clone --recursive git@github.com:digiu-ai/p2p-bridge.git
cd p2p-bridge
./deploy.sh
```


### запуск теста
 
```
make -C external/eth-contracts local-test
 
```

