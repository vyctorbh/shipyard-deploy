#!/bin/bash
DB_PASS=${DB_PASS:-1q2w3e4r5t}
ADMIN_PASS=${ADMIN_PASS:-shipyard}
ACTION=${1:-}
if [ ! -e "/docker.sock" ] ; then
    echo "You must map your Docker socket to /docker.sock (i.e. -v /var/run/docker.sock:/docker.sock)"
    exit 2
fi

if [ -z "$ACTION" ] ; then
    echo "Usage: "
    echo "  shipyard/deploy <action>"
    echo ""
    echo "Examples: "
    echo "  docker run shipyard/deploy setup"
    echo "     Deploys Shipyard stack"
    echo "  docker run shipyard/deploy cleanup"
    echo "     Removes Shipyard stack"
    exit 1
fi

if [ "$ACTION" = "setup" ] ; then
    set -e
    docker -H unix://docker.sock run -i -t -d -p 6379 -name shipyard_redis shipyard/redis
    docker -H unix://docker.sock run -i -t -d -p 80 -link shipyard_redis:redis -name shipyard_router shipyard/router
    docker -H unix://docker.sock run -i -t -d -p 80:80 -link shipyard_redis:redis -link shipyard_router:app_router -name shipyard_lb shipyard/lb
    docker -H unix://docker.sock run -i -t -d -p 5432 -e DB_PASS=$DB_PASS -name shipyard_db shipyard/db
    docker -H unix://docker.sock run -i -t -d -p 8000:8000 -name shipyard -e ADMIN_PASS=$ADMIN_PASS shipyard/shipyard app master-worker
elif [ "$ACTION" = "cleanup" ] ; then
    for CNT in shipyard shipyard_redis shipyard_router shipyard_db shipyard_lb
    do
        docker -H unix://docker.sock kill $CNT > /dev/null 2>&1
        docker -H unix://docker.sock rm $CNT > /dev/null 2>&1
    done
fi


