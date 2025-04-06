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

# Проект второго задания, ч. 3

Лог запуска pg-connector:
```
2025-04-06 22:34:19 [2025-04-06 19:34:19,646] INFO Kafka version: 7.7.1-ccs (org.apache.kafka.common.utils.AppInfoParser)
2025-04-06 22:34:19 [2025-04-06 19:34:19,646] INFO Kafka commitId: 91d86f33092378c89731b4a9cf1ce5db831a2b07 (org.apache.kafka.common.utils.AppInfoParser)
2025-04-06 22:34:19 [2025-04-06 19:34:19,646] INFO Kafka startTimeMs: 1743968059646 (org.apache.kafka.common.utils.AppInfoParser)
2025-04-06 22:34:19 [2025-04-06 19:34:19,647] INFO [Producer clientId=connector-producer-pg-connector-0] Cluster ID: practicum (org.apache.kafka.clients.Metadata)
2025-04-06 22:34:19 [2025-04-06 19:34:19,656] INFO [Worker clientId=connect-localhost:8083, groupId=kafka-connect] Finished starting connectors and tasks (org.apache.kafka.connect.runtime.distributed.DistributedHerder)
2025-04-06 22:34:19 [2025-04-06 19:34:19,657] INFO Starting PostgresConnectorTask with configuration:
2025-04-06 22:34:19    connector.class = io.debezium.connector.postgresql.PostgresConnector
2025-04-06 22:34:19    database.user = postgres-user
2025-04-06 22:34:19    database.dbname = customers
2025-04-06 22:34:19    transforms.unwrap.delete.handling.mode = rewrite
2025-04-06 22:34:19    topic.creation.default.partitions = -1
2025-04-06 22:34:19    transforms = unwrap
2025-04-06 22:34:19    database.server.name = customers
2025-04-06 22:34:19    database.port = 5432
2025-04-06 22:34:19    table.whitelist = public.customers
2025-04-06 22:34:19    topic.creation.enable = true
2025-04-06 22:34:19    topic.prefix = customers
2025-04-06 22:34:19    task.class = io.debezium.connector.postgresql.PostgresConnectorTask
2025-04-06 22:34:19    database.hostname = postgres
2025-04-06 22:34:19    database.password = ********
2025-04-06 22:34:19    transforms.unwrap.drop.tombstones = false
2025-04-06 22:34:19    topic.creation.default.replication.factor = -1
2025-04-06 22:34:19    name = pg-connector
2025-04-06 22:34:19    transforms.unwrap.type = io.debezium.transforms.ExtractNewRecordState
2025-04-06 22:34:19    skipped.operations = none
2025-04-06 22:34:19  (io.debezium.connector.common.BaseSourceTask)
2025-04-06 22:34:19 [2025-04-06 19:34:19,658] INFO Loading the custom source info struct maker plugin: io.debezium.connector.postgresql.PostgresSourceInfoStructMaker (io.debezium.config.CommonConnectorConfig)
2025-04-06 22:34:19 [2025-04-06 19:34:19,659] INFO Loading the custom topic naming strategy plugin: io.debezium.schema.SchemaTopicNamingStrategy (io.debezium.config.CommonConnectorConfig)
2025-04-06 22:34:19 [2025-04-06 19:34:19,668] INFO Connection gracefully closed (io.debezium.jdbc.JdbcConnection)
2025-04-06 22:34:19 [2025-04-06 19:34:19,707] WARN Type [oid:13529, name:_pg_user_mappings] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,707] WARN Type [oid:13203, name:cardinal_number] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,707] WARN Type [oid:13206, name:character_data] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,707] WARN Type [oid:13208, name:sql_identifier] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,707] WARN Type [oid:13214, name:time_stamp] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,707] WARN Type [oid:13216, name:yes_or_no] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,745] INFO No previous offsets found (io.debezium.connector.common.BaseSourceTask)
2025-04-06 22:34:19 [2025-04-06 19:34:19,762] WARN Type [oid:13529, name:_pg_user_mappings] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,762] WARN Type [oid:13203, name:cardinal_number] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,762] WARN Type [oid:13206, name:character_data] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,762] WARN Type [oid:13208, name:sql_identifier] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,762] WARN Type [oid:13214, name:time_stamp] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,762] WARN Type [oid:13216, name:yes_or_no] is already mapped (io.debezium.connector.postgresql.TypeRegistry)
2025-04-06 22:34:19 [2025-04-06 19:34:19,799] INFO Connector started for the first time. (io.debezium.connector.common.BaseSourceTask)
2025-04-06 22:34:19 [2025-04-06 19:34:19,800] INFO No previous offset found (io.debezium.connector.postgresql.PostgresConnectorTask)
2025-04-06 22:34:19 [2025-04-06 19:34:19,806] INFO user 'postgres-user' connected to database 'customers' on PostgreSQL 16.4 (Debian 16.4-1.pgdg110+2) on x86_64-pc-linux-gnu, compiled by gcc (Debian 10.2.1-6) 10.2.1 20210110, 64-bit with roles:
2025-04-06 22:34:19     role 'pg_read_all_settings' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_database_owner' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_stat_scan_tables' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_checkpoint' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_write_server_files' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_use_reserved_connections' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'postgres-user' [superuser: true, replication: true, inherit: true, create role: true, create db: true, can log in: true]
2025-04-06 22:34:19     role 'pg_read_all_data' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_write_all_data' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_monitor' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_read_server_files' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_create_subscription' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_execute_server_program' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_read_all_stats' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false]
2025-04-06 22:34:19     role 'pg_signal_backend' [superuser: false, replication: false, inherit: true, create role: false, create db: false, can log in: false] (io.debezium.connector.postgresql.PostgresConnectorTask)
2025-04-06 22:34:19 [2025-04-06 19:34:19,813] INFO Obtained valid replication slot ReplicationSlot [active=false, latestFlushedLsn=null, catalogXmin=null] (io.debezium.connector.postgresql.connection.PostgresConnection)
2025-04-06 22:34:19 [2025-04-06 19:34:19,829] INFO Creating replication slot with command CREATE_REPLICATION_SLOT "debezium"  LOGICAL decoderbufs  (io.debezium.connector.postgresql.connection.PostgresReplicationConnection)
2025-04-06 22:34:19 [2025-04-06 19:34:19,897] INFO Requested thread factory for component PostgresConnector, id = customers named = SignalProcessor (io.debezium.util.Threads)
2025-04-06 22:34:19 [2025-04-06 19:34:19,914] INFO Requested thread factory for component PostgresConnector, id = customers named = change-event-source-coordinator (io.debezium.util.Threads)
2025-04-06 22:34:19 [2025-04-06 19:34:19,914] INFO Requested thread factory for component PostgresConnector, id = customers named = blocking-snapshot (io.debezium.util.Threads)
2025-04-06 22:34:19 [2025-04-06 19:34:19,920] INFO Creating thread debezium-postgresconnector-customers-change-event-source-coordinator (io.debezium.util.Threads)
2025-04-06 22:34:19 [2025-04-06 19:34:19,920] INFO WorkerSourceTask{id=pg-connector-0} Source task finished initialization and start (org.apache.kafka.connect.runtime.AbstractWorkerSourceTask)
2025-04-06 22:34:19 [2025-04-06 19:34:19,923] INFO Metrics registered (io.debezium.pipeline.ChangeEventSourceCoordinator)
2025-04-06 22:34:19 [2025-04-06 19:34:19,923] INFO Context created (io.debezium.pipeline.ChangeEventSourceCoordinator)
2025-04-06 22:34:19 [2025-04-06 19:34:19,928] INFO According to the connector configuration data will be snapshotted (io.debezium.connector.postgresql.PostgresSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,932] INFO Snapshot step 1 - Preparing (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,933] INFO Setting isolation level (io.debezium.connector.postgresql.PostgresSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,933] INFO Opening transaction with statement SET TRANSACTION ISOLATION LEVEL REPEATABLE READ; 
2025-04-06 22:34:19 SET TRANSACTION SNAPSHOT '00000006-00000002-1'; (io.debezium.connector.postgresql.PostgresSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,989] INFO Snapshot step 2 - Determining captured tables (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,991] INFO Adding table public.users to the list of capture schema tables (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,992] INFO Created connection pool with 1 threads (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,992] INFO Snapshot step 3 - Locking captured tables [public.users] (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,995] INFO Snapshot step 4 - Determining snapshot offset (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:19 [2025-04-06 19:34:19,997] INFO Creating initial offset context (io.debezium.connector.postgresql.PostgresOffsetContext)
2025-04-06 22:34:20 [2025-04-06 19:34:19,999] INFO Read xlogStart at 'LSN{0/1A43568}' from transaction '745' (io.debezium.connector.postgresql.PostgresOffsetContext)
2025-04-06 22:34:20 [2025-04-06 19:34:20,007] INFO Read xlogStart at 'LSN{0/1A43568}' from transaction '745' (io.debezium.connector.postgresql.PostgresSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,007] INFO Snapshot step 5 - Reading structure of captured tables (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,008] INFO Reading structure of schema 'public' of catalog 'customers' (io.debezium.connector.postgresql.PostgresSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,028] INFO Snapshot step 6 - Persisting schema history (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,028] INFO Snapshot step 7 - Snapshotting data (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,030] INFO Creating snapshot worker pool with 1 worker thread(s) (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,032] INFO For table 'public.users' using select statement: 'SELECT "id", "name", "private_info" FROM "public"."users"' (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,035] INFO Exporting data from table 'public.users' (1 of 1 tables) (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,051] INFO       Finished exporting 3 records for table 'public.users' (1 of 1 tables); total duration '00:00:00.016' (io.debezium.relational.RelationalSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,053] INFO Snapshot - Final stage (io.debezium.pipeline.source.AbstractSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,054] INFO Snapshot completed (io.debezium.pipeline.source.AbstractSnapshotChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,059] INFO Snapshot ended with SnapshotResult [status=COMPLETED, offset=PostgresOffsetContext [sourceInfoSchema=Schema{io.debezium.connector.postgresql.Source:STRUCT}, sourceInfo=source_info[server='customers'db='customers', lsn=LSN{0/1A43568}, txId=745, timestamp=2025-04-06T19:34:19.935456Z, snapshot=FALSE, schema=public, table=users], lastSnapshotRecord=true, lastCompletelyProcessedLsn=null, lastCommitLsn=null, streamingStoppingLsn=null, transactionContext=TransactionContext [currentTransactionId=null, perTableEventCount={}, totalEventCount=0], incrementalSnapshotContext=IncrementalSnapshotContext [windowOpened=false, chunkEndPosition=null, dataCollectionsToSnapshot=[], lastEventKeySent=null, maximumKey=null]]] (io.debezium.pipeline.ChangeEventSourceCoordinator)
2025-04-06 22:34:20 [2025-04-06 19:34:20,061] INFO Connected metrics set to 'true' (io.debezium.pipeline.ChangeEventSourceCoordinator)
2025-04-06 22:34:20 [2025-04-06 19:34:20,079] INFO REPLICA IDENTITY for 'public.users' is 'DEFAULT'; UPDATE and DELETE events will contain previous values only for PK columns (io.debezium.connector.postgresql.PostgresSchema)
2025-04-06 22:34:20 [2025-04-06 19:34:20,090] INFO SignalProcessor started. Scheduling it every 5000ms (io.debezium.pipeline.signal.SignalProcessor)
2025-04-06 22:34:20 [2025-04-06 19:34:20,090] INFO Creating thread debezium-postgresconnector-customers-SignalProcessor (io.debezium.util.Threads)
2025-04-06 22:34:20 [2025-04-06 19:34:20,090] INFO Starting streaming (io.debezium.pipeline.ChangeEventSourceCoordinator)
2025-04-06 22:34:20 [2025-04-06 19:34:20,090] INFO Retrieved latest position from stored offset 'LSN{0/1A43568}' (io.debezium.connector.postgresql.PostgresStreamingChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,092] INFO Looking for WAL restart position for last commit LSN 'null' and last change LSN 'LSN{0/1A43568}' (io.debezium.connector.postgresql.connection.WalPositionLocator)
2025-04-06 22:34:20 [2025-04-06 19:34:20,101] INFO Obtained valid replication slot ReplicationSlot [active=false, latestFlushedLsn=LSN{0/1A43568}, catalogXmin=745] (io.debezium.connector.postgresql.connection.PostgresConnection)
2025-04-06 22:34:20 [2025-04-06 19:34:20,102] INFO Connection gracefully closed (io.debezium.jdbc.JdbcConnection)
2025-04-06 22:34:20 [2025-04-06 19:34:20,120] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:20 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 2 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:20 [2025-04-06 19:34:20,120] INFO Requested thread factory for component PostgresConnector, id = customers named = keep-alive (io.debezium.util.Threads)
2025-04-06 22:34:20 [2025-04-06 19:34:20,121] INFO Creating thread debezium-postgresconnector-customers-keep-alive (io.debezium.util.Threads)
2025-04-06 22:34:20 [2025-04-06 19:34:20,136] INFO REPLICA IDENTITY for 'public.users' is 'DEFAULT'; UPDATE and DELETE events will contain previous values only for PK columns (io.debezium.connector.postgresql.PostgresSchema)
2025-04-06 22:34:20 [2025-04-06 19:34:20,136] INFO Processing messages (io.debezium.connector.postgresql.PostgresStreamingChangeEventSource)
2025-04-06 22:34:20 [2025-04-06 19:34:20,428] INFO The task will send records to topic 'customers.public.users' for the first time. Checking whether topic exists (org.apache.kafka.connect.runtime.AbstractWorkerSourceTask)
2025-04-06 22:34:20 [2025-04-06 19:34:20,431] INFO Creating topic 'customers.public.users' (org.apache.kafka.connect.runtime.AbstractWorkerSourceTask)
2025-04-06 22:34:20 [2025-04-06 19:34:20,476] INFO Created topic (name=customers.public.users, numPartitions=-1, replicationFactor=-1, replicasAssignments=null, configs={}) on brokers at kafka-0:9092 (org.apache.kafka.connect.util.TopicAdmin)
2025-04-06 22:34:20 [2025-04-06 19:34:20,476] INFO Created topic '(name=customers.public.users, numPartitions=-1, replicationFactor=-1, replicasAssignments=null, configs={})' using creation group TopicCreationGroup{name='default', inclusionPattern=.*, exclusionPattern=, numPartitions=-1, replicationFactor=-1, otherConfigs={}} (org.apache.kafka.connect.runtime.AbstractWorkerSourceTask)
2025-04-06 22:34:20 [2025-04-06 19:34:20,481] WARN [Producer clientId=connector-producer-pg-connector-0] Got error produce response with correlation id 6 on topic-partition customers.public.users-0, retrying (2147483646 attempts left). Error: UNKNOWN_TOPIC_OR_PARTITION (org.apache.kafka.clients.producer.internals.Sender)
2025-04-06 22:34:20 [2025-04-06 19:34:20,482] WARN [Producer clientId=connector-producer-pg-connector-0] Received unknown topic or partition error in produce request on partition customers.public.users-0. The topic-partition may not exist or the user may not have Describe access to it (org.apache.kafka.clients.producer.internals.Sender)
2025-04-06 22:34:20 [2025-04-06 19:34:20,578] WARN [Producer clientId=connector-producer-pg-connector-0] Got error produce response with correlation id 8 on topic-partition customers.public.users-0, retrying (2147483645 attempts left). Error: UNKNOWN_TOPIC_OR_PARTITION (org.apache.kafka.clients.producer.internals.Sender)
2025-04-06 22:34:20 [2025-04-06 19:34:20,578] WARN [Producer clientId=connector-producer-pg-connector-0] Received unknown topic or partition error in produce request on partition customers.public.users-0. The topic-partition may not exist or the user may not have Describe access to it (org.apache.kafka.clients.producer.internals.Sender)
2025-04-06 22:34:20 [2025-04-06 19:34:20,804] WARN [Producer clientId=connector-producer-pg-connector-0] Got error produce response with correlation id 10 on topic-partition customers.public.users-0, retrying (2147483644 attempts left). Error: UNKNOWN_TOPIC_OR_PARTITION (org.apache.kafka.clients.producer.internals.Sender)
2025-04-06 22:34:20 [2025-04-06 19:34:20,804] WARN [Producer clientId=connector-producer-pg-connector-0] Received unknown topic or partition error in produce request on partition customers.public.users-0. The topic-partition may not exist or the user may not have Describe access to it (org.apache.kafka.clients.producer.internals.Sender)
2025-04-06 22:34:25 [2025-04-06 19:34:25,150] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:25 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:30 [2025-04-06 19:34:30,183] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:30 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:35 [2025-04-06 19:34:35,219] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:35 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:40 [2025-04-06 19:34:40,251] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:40 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 2 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:42 [2025-04-06 19:34:42,023] WARN Ignoring invalid completion time 1743968082022 since it is before this stage's start time of 1743968082538 (org.apache.kafka.connect.util.Stage)
2025-04-06 22:34:41 [2025-04-06 19:34:41,450] WARN Ignoring invalid completion time 1743968081450 since it is before this stage's start time of 1743968082023 (org.apache.kafka.connect.util.Stage)
2025-04-06 22:34:43 [2025-04-06 19:34:43,987] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:43 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:49 [2025-04-06 19:34:49,018] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:49 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:54 [2025-04-06 19:34:54,050] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:54 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:34:59 [2025-04-06 19:34:59,079] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:34:59 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:04 [2025-04-06 19:35:04,110] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:35:04 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:09 [2025-04-06 19:35:09,139] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:35:09 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:10 [2025-04-06 19:35:10,718] WARN Ignoring invalid completion time 1743968110718 since it is before this stage's start time of 1743968111223 (org.apache.kafka.connect.util.Stage)
2025-04-06 22:35:10 [2025-04-06 19:35:10,145] WARN Ignoring invalid completion time 1743968110145 since it is before this stage's start time of 1743968110718 (org.apache.kafka.connect.util.Stage)
2025-04-06 22:35:12 [2025-04-06 19:35:12,893] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:35:12 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:14 [2025-04-06 19:35:14,371] INFO 172.18.0.9 - - [06/Apr/2025:19:35:14 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "Apache-HttpClient/4.5.14 (Java/11.0.21)" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:14 [2025-04-06 19:35:14,399] INFO 172.18.0.9 - - [06/Apr/2025:19:35:14 +0000] "GET /connectors/pg-connector HTTP/1.1" 200 778 "-" "Apache-HttpClient/4.5.14 (Java/11.0.21)" 5 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:14 [2025-04-06 19:35:14,556] INFO 172.18.0.9 - - [06/Apr/2025:19:35:14 +0000] "GET /connectors/pg-connector/status HTTP/1.1" 200 166 "-" "Apache-HttpClient/4.5.14 (Java/11.0.21)" 4 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:17 [2025-04-06 19:35:17,088] INFO WorkerSourceTask{id=pg-connector-0} Committing offsets for 3 acknowledged messages (org.apache.kafka.connect.runtime.WorkerSourceTask)
2025-04-06 22:35:17 [2025-04-06 19:35:17,100] INFO Found previous partition offset PostgresPartition [sourcePartition={server=customers}]: {lsn=27538792, txId=745, ts_usec=1743968059935456} (io.debezium.connector.common.BaseSourceTask)
2025-04-06 22:35:17 [2025-04-06 19:35:17,924] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:35:17 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:22 [2025-04-06 19:35:22,956] INFO [0:0:0:0:0:0:0:1] - - [06/Apr/2025:19:35:22 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "curl/7.61.1" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:24 [2025-04-06 19:35:24,574] INFO 172.18.0.9 - - [06/Apr/2025:19:35:24 +0000] "GET /connectors HTTP/1.1" 200 16 "-" "Apache-HttpClient/4.5.14 (Java/11.0.21)" 1 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:24 [2025-04-06 19:35:24,579] INFO 172.18.0.9 - - [06/Apr/2025:19:35:24 +0000] "GET /connectors/pg-connector HTTP/1.1" 200 778 "-" "Apache-HttpClient/4.5.14 (Java/11.0.21)" 2 (org.apache.kafka.connect.runtime.rest.RestServer)
2025-04-06 22:35:24 [2025-04-06 19:35:24,585] INFO 172.18.0.9 - - [06/Apr/2025:19:35:24 +0000] "GET /connectors/pg-connector/status HTTP/1.1" 200 166 "-" "Apache-HttpClient/4.5.14 (Java/11.0.21)" 3 (org.apache.kafka.connect.runtime.rest.RestServer)
```
