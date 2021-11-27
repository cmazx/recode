# recode
Distributed scalable media process service

1 Installation
Firstly you should install docker, docker-compose and  traefik revese proxy container locally.

Install docker
https://docs.docker.com/engine/install/

Install docker-compose
https://docs.docker.com/compose/install/

Run reverse-proxy
```
docker network create router

docker run --detach --restart=always --name router --network=router \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -p 80:80 \
    -l traefik.enable=true \
    -l traefik.frontend.rule=Host:router.box \
    -l traefik.port=8080 \
    traefik:v1.7.15-alpine --api --docker --docker.exposedbydefault=false
```

2 Add hosts to your /etc/hosts file ( or c:/windows/system32/drivers/etc/hosts)

127.0.0.1 worker.recode.box

127.0.0.1 api.recode.box

127.0.0.1 kafka.recode.box

127.0.0.1 storage.recode.box
