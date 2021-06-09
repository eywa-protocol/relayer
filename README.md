## децентрализованный релеер для выполнения кросc-чейн вызовов


#### пререквизиты
 go, truffle, npx


###  локальный деплой

```bash
mkdir digiu
cd digiu
git clone --recursive git@github.com:digiu-ai/p2p-bridge.git
cd p2p-bridge/scripts
# for macos only run before deploy sudo ./macos_add_interfaces.sh
./deploy.sh
```


### запуск теста
 
```
make -C external/eth-contracts local-test
 
```

