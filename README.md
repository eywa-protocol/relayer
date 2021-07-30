## децентрализованный релеер для выполнения кросc-чейн вызовов

#### пререквизиты

go, hardhat, npx, docker, docker-compose

### локальный деплой

```shell
mkdir digiu
cd digiu
git clone --recursive git@gitlab.digiu.ai:blockchainlaboratory/eywa-p2p-bridge.git
cd p2p-bridge
git submodule foreach -q --recursive 'git checkout $(git config -f $toplevel/.gitmodules submodule.$name.branch || echo main)'
make -C external/eth-contracts/
make
cd scripts

# for macos only run before deploy sudo ./macos_add_interfaces.sh
./deploy.sh
```

### запуск теста

```shell
make -C external/eth-contracts local-test
 
```

### Переключение loglevel

Переключение loglevel между заданным при старте и trace выполняется отправкой сигнала USR2 процессу

Например для node_1

```shell
docker-compose exec --index=1 node kill -12 1
```

### Ребилд нод после изменения кода

```shell
cd scripts
./rebuild_nodes.sh
```

### Остановка всех контейнеров

```shell
cd scripts
./stop_all.sh
```

### Запуск всех контейнеров

```shell
cd scripts
./start_all.sh

```

### Запуск теста

```shell
for (( c=0; c<=10; c++ )) do go test -count=1 test/testnets_test.go; done
```

### Sentry
Для включения отправки ошибок в Sentry выполните ```./init_env_sentry.sh``` из папки скрипт с параметром имя разработчика

Пример
```shell
cd scripts
./init_env_sentry.sh dkh
```

