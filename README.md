# Исмагилов Тимур Backend
Выполнены все подзадания. Есть Swagger-документ. Настроено Docker-окружение для деплоя и тестирования. Использована база данных PostgreSQL.

## Вопросы
* Не сказано, какие могут быть id пользователей. Предполагаю, что id — любое целое число.
* Не сказано, какие могут быть названия сегментов. Предполагаю, что любые непустые строки.
* Не сказано, можно ли использовать кодогенерацию на основе Swagger-файла. Решил, что можно, потому что это сокращает time-to-market и количество ошибок. Использовал swagger-codegen. Но потом мне пришлось рефакторить автосгенерированный код из-за шумности. В конечном коде от кодгена не осталось почти ничего.
* Что делать, если создаётся сегмент с названием, как у сегмента, который был когда-то удалён? Я решил, что так нельзя, пускай аналитики из отдела каждый раз уникальное название для экспериментов придумывают. Так и выгрузка отчётов однозначнее.
* Не очень понятно, что значит, что какой-то процент пользователей будет попадать в сегмент автоматически. Должны ли старые известные пользователи попадать в этот процент ретроспективно? Или только новые? Надо ли учитывать явно добавленных в этот процент? Насколько строго должен этот процент держаться?
 
  Я решил так. Когда создаётся сегмент с процентом _n_, _n_ процентам известным ранее пользователям ретроспективно задаётся такой сегмент разово. Далее, при появлении информации о ранее неизвестном пользователе, этот пользователь с вероятностью _n_ будет добавлен в сегмент.

## Запуск
Требуется Docker.

### Как прогнать тесты
```shell
make test
```

### Как запустить сервер
```shell
make run
```

Делайте запросы к [localhost:8080](http://localhost:8080).

### Как сбросить базу данных
```shell
make clear
```

## Примеры API
Для полного описания API см. файл `swagger.yaml`. Ниже будут приведены некоторые примеры для представления.

### Создать сегмент
```shell
curl http://localhost:8080/create_segment -X POST -H 'Content-Type: application/json'\
  -d '{"name":"BOUNCEPAW_SEGMENT","percent":30}'
```

### Удалить сегмент
```shell
curl http://localhost:8080/delete_segment -X POST -H 'Content-Type: application/json'\
  -d '{"name":"BOUNCEPAW_SEGMENT"}'
```

### Обновить данные пользователя
```shell
curl http://localhost:8080/update_user -X POST -H 'Content-Type: application/json'\
  -d '{"id":1000,"add_to_segments":["BOUNCEPAW_SEGMENT"],"remove_from_segments":["AVITO_SEGMENT"],"ttl":86400}'
```
Примечание: в одном дне 86400 секунд. TTL указывается в секундах.

### Получить данные о пользователе
```shell
curl http://localhost:8080/get_segments -X POST -H 'Content-Type: application/json'\
  -d '{"id":1000}'
```

### Узнать адрес файла с историей операций за этот месяц
Допустим, что сегодня сентябрь 2023.
```shell
curl http://localhost:8080/history -X POST -H 'Content-Type: application/json'\
  -d '{"year":2023,"month":9}'
```

Берём адрес из поля `link` ответа.

### Загрузить данные об операциях за месяц
```shell
curl 'http://localhost:8080/history?year=2023&month=9' -X GET
```
