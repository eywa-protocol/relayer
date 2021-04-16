Это демо-версия для тестирования dht

Содержание теста：
Используя публичные узлы ipfs в качестве bootstrapPeer, несколько узлов могут обнаруживать друг друга и подключаться друг к другу

Создайте узел bootstrapPeer локально для достижения взаимного обнаружения и соединения нескольких узлов

Введение в документ：
начальная загрузка.go работает на локальном узле начальной загрузки
норма.перейти к обычному узлу, после успешного подключения к загрузочному узлу, вы также можете стать загрузочным узлом
применение：
Использование общедоступных узлов в качестве bootstrapPeer：

### Сначала запустите узел, чтобы создать комнату

``
go run normPeer.go -room myroom
``

### Запустите второй узел в другом терминале

``
go run normPeer.go -joinRoom myroom
``



Использование локально созданного bootstrapPeer

### Используйте buildw.сборка скрипта bat
 ``
. ./build.sh 
``



### Сначала запустите узел начальной загрузки
``
go run bootstrap.go
``
### Выведите следующее：
``
Addr:/ip4/10.0.0.193/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
``
``
Addr:/ip4/127.0.0.1/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
``
### Эти два адреса предназначены для подключения других узлов norm


### Запустите другой узел и создайте комнату

``
go run normPeer.go -room myroom -b /ip4/10.0.0.193/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
``

### Запустите другой узел в другом терминале

``
go run normPeer.go -joinRoom myroom -b /ip4/10.0.0.193/tcp/6666/p2p/QmeYXhotakHDNtZcvZzz9prWp2HY3wNEPMzTRojV1FCkdk
``
