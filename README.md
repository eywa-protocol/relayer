## децентрализованный релеер для выполнения кросc-чейн вызовов

#### пререквизиты

go, hardhat, npx, docker, docker-compose


### настройка репозитория
```shell
mkdir digiu
cd digiu
git clone git@gitlab.digiu.ai:blockchainlaboratory/eywa-p2p-bridge.git
cd eywa-p2p-bridge
```
Добавление submodule (если необходимо)
```shell
git submodule init
git submodule update (--remote - если необходимо получить последние изменения)
```

### тестнет деплой
- Рекомендуется предварительно очищать каталог external/eth-contracts/hardhat/networks_env, если он существует (возможны конфликты с настройками нод)
- Адрес релеера во всех всетях 0x2b3cc5fcAC62299520FA96D75f125c33B48E70d7

```shell
make -C external/eth-contracts eth-testnet-migrate
cd scripts
./deploy.sh testnet
```

### локальный деплой
- Рекомендуется предварительно очищать каталог external/eth-contracts/hardhat/networks_env, если он существует (возможны конфликты с настройками нод)

```shell
cd scripts

# for macos only run before deploy sudo ./macos_add_interfaces.sh
./deploy.sh local
```

### запуск тестов

- e2e тест js
```
make -C external/eth-contracts testnet-test
make -C external/eth-contracts local-test
```
- e2e тест golang
```
./run_go_test.sh testnet
./run_go_test.sh local
```

### Работа с репозиторием
Чтобы зафиксировать версию подмодуля в родительском репозитории:
```
cd external/eth-contracts
git checkout develop (или другая необходимая ветка)
*ВНОCИМ ИЗМЕНЕНИЯ*
git add .
git commit -m "commit message"
git push
cd ../../
git submodule update --remote
git add external/eth-contracts
git commit -m "submodule eth-contracts update"
git push
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



