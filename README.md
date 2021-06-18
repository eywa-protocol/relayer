## децентрализованный релеер для выполнения кросc-чейн вызовов


#### пререквизиты
 go, truffle, npx, docker, docker-compose


###  локальный деплой

```bash
mkdir digiu
cd digiu
git clone --recursive git@github.com:digiu-ai/p2p-bridge.git
cd p2p-bridge
git submodule foreach -q --recursive 'git checkout $(git config -f $toplevel/.gitmodules submodule.$name.branch || echo main)'
make -C external/eth-contracts/
make
cd scripts
# for macos only run before deploy sudo ./macos_add_interfaces.sh
./deploy.sh
```


### запуск теста
 
```
make -C external/eth-contracts local-test
 
```

### Переключение loglevel

Переключение loglevel между заданным при старте и trace выполняется отправкой сигнала USR2 процессу 

Например для node_1
```
docker-compose exec --index=1 node kill -12 1
```
