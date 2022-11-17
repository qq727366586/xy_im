# xy_im


-- goim部署文档
version: '2'
services:
redis:
image: redis
ports:
- "6379:6379"
networks:
-  dev
zookeeper:
image: wurstmeister/zookeeper   ## 镜像
ports:
- "2181:2181"                 ## 对外暴露的端口号
networks:
-  dev
kafka:
image: wurstmeister/kafka       ## 镜像
volumes:
- /etc/localtime:/etc/localtime ## 挂载位置（kafka镜像和宿主机器之间时间保持一直）
ports:
- "9094:9092"
networks:
-  dev
environment:
KAFKA_BROKER_ID=1
KAFKA_ADVERTISED_HOST_NAME: 172.16.0.86  ## 修改:宿主机IP
KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181       ## 卡夫卡运行是基于zookeeper的
KAFKA_ADVERTISED_PORT: 9094
KAFKA_LOG_RETENTION_HOURS: 120
KAFKA_MESSAGE_MAX_BYTES: 10000000
KAFKA_REPLICA_FETCH_MAX_BYTES: 10000000
KAFKA_GROUP_MAX_SESSION_TIMEOUT_MS: 60000
KAFKA_NUM_PARTITIONS: 3
KAFKA_DELETE_RETENTION_MS: 1000
kafka-manager:  
image: sheepkiller/kafka-manager                ## 镜像：开源的web管理kafka集群的界面
environment:
ZK_HOSTS: 172.16.0.86                       ## 修改:宿主机IP
ports:  
- "9001:9000"                                 ## 暴露端口
networks:
-  dev
networks:
dev:
driver: bridge

   
kafka-topics.sh --create --topic goim-push-topic --bootstrap-server 172.16.0.86:9094 --replication-factor 1 --partitions 4

go run main.go -conf discovery.toml

go run main.go -conf=comet-example.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 -addrs=127.0.0.1 -debug=true
go run main.go -conf=job-example.toml -region=sh -zone=sh001 -deploy.env=dev
go run main.go -conf=logic-example.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10

