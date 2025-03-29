# Итоговый проект третьего модуля

## Проект второго задания, ч. 1

- poll.interval.ms=200
- Эксперименты 2-8 имеют параметры batch.max.rows для 1000 записей размером 528 байт
- Эксперименты 9-13 имеют параметры batch.max.rows для 2000 записей размером 528 байт

| Эксперимент  | batch.size | linger.ms | compression.type | buffer.memory | Source Record Write Rate (кops/sec) |
| ------------- | ------------- | ------------- | ------------- | ------------- | ------------- |
| 1  | 100  | 0  | none  | 33554432  | 7.22kops |
| 2  | 528000  | 0  | none  | 33554432  | 156kops |
| 3  | 528000  | 500  | none  | 33554432  | 162kops |
| 4  | 528000  | 1000  | none  | 33554432  | 152kops |
| 5  | 528000  | 500  | gzip  | 33554432  | 161kops |
| 6  | 528000  | 1000  | gzip  | 33554432  | 153kops |
| 7  | 528000  | 500  | snappy  | 33554432  | 157kops |
| 8  | 528000  | 1000  | snappy  | 33554432  | 160kops |
| 9  | 1056000  | 500  | none  | 33554432  | 0kops (Kafka fail: MESSAGE_TOO_LARGE) |
| 10 | 1056000  | 500  | gzip  | 33554432  | 150kops  |
| 11 | 1056000  | 1000  | gzip  | 33554432  | 149kops  |
| 12  | 1056000  | 500  | snappy  | 33554432  | 162kops  |
| 13  | 1056000  | 1000  | snappy  | 33554432  | 152kops  |

### Анализ экспериментов

- Отправка маленьких сообщений малыми порциями идёт слишком долго.
- Необходимо время, чтобы коннектор успел набрать достаточное количество данных для отправки. Для этого надо увеличить linger.ms.
- Слишком большие пачки данных могут не пройти ограничения брокера (см. эксперимент 9).
- Слишком большие пачки данных дольше сжимать (см. эксперименты 5 vs 10, 12 vs 13), при этом это может повлиять на производительность.
- Snappy действительно позволяет лучшую производительность, сжимая данные против gzip.
- Увеличение буфера, объема данных для передачи, при должных параметрах увеличивает производительность (см. эксперименты 7 vs 12).
- Сжатые данные позволяют экономить место в топике в разы относительно несжатых данных.

## Проект второго задания, ч. 2

- Собрать кастомный сервис-коннектор `docker build -t go-app ./src/.`
- Перейти в папку __infra__ и выполнить `docker compose up -d`
- Создать топик для отправки метрик `docker exec -it infra-kafka-0-1 /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic metrics --bootstrap-server localhost:9092 --partitions 3 --replication-factor 2`

### Компоненты
- x-kafka-common - общая информация контейнеров kafka-*
- kafka-0, kafka-1, kafka-2 - контейнеры Kafka cluster
- ui - Kafka UI для управления кластером
    - Зависит от kafka-*
- my-service - кастомный сервис-коннектор
    - Зависит от kafka-*
    - Конфигурация описана внутри `docker-compose.yaml`
- prometheus - контейнер с Prometheus
    - Поднимается независимо
    - Связан с my-service в контексте 2 задания
- Описание остальных компонент опущено, т.к. они вне 2 задания

### Сценарий тестирования

1. Перейти на [UI Kafka](http://localhost:8085/ui/clusters/kraft/all-topics?perPage=25)
2. Зайти в топик `metrics`
3. Отправить сообщение через Produce Message
    - Пример сообщения:
```
{
    "Alloc": {
        "Type": "gauge",
        "Name": "Alloc",
        "Description": "Alloc is bytes of allocated heap objects.",
        "Value": 24293912
    },
    "FreeMemory": {
        "Type": "gauge",
        "Name": "FreeMemory",
        "Description": "RAM available for programs to allocate",
        "Value": 7740977152
    },
    "PollCount": {
        "Type": "counter",
        "Name": "PollCount",
        "Description": "PollCount is quantity of metrics collection iteration.",
        "Value": 3
    },
    "TotalMemory": {
        "Type": "gauge",
        "Name": "TotalMemory",
        "Description": "Total amount of RAM on this system",
        "Value": 16054480896
    }
}
```
4. Зайти на [UI Prometheus](http://localhost:9090/classic/graph)
    - Нажать `insert metric at cursor`
    - Выбрать метрику `Alloc`, например
    - Сделать то же самое для оставшихся метрик: FreeMemory, PollCount, TotalMemory
5. Отправить сообщение через Produce Message в Kafka UI, изменив метрики
6. Увидеть, что изменения пролились в Prometheus

