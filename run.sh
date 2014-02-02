#!/bin/bash
DB_PASS=${DB_PASS:-1q2w3e4r5t}
ADMIN_PASS=${ADMIN_PASS:-shipyard}
TAG=${TAG:-latest}
AGENT_VERSION=${AGENT_VERSION:-v0.0.9}
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

function cleanup {
    for CNT in shipyard shipyard_redis shipyard_router shipyard_db shipyard_lb shipyard_agent
    do
        docker -H unix://docker.sock kill $CNT > /dev/null 2>&1
        docker -H unix://docker.sock rm $CNT > /dev/null 2>&1
    done
}

function purge {
    cleanup
    for IMG in shipyard redis router db lb agent
    do
        docker -H unix://docker.sock rmi shipyard/$IMG > /dev/null 2>&1
    done
}

if [ "$ACTION" = "setup" ] ; then
    echo "Using $TAG tag for Shipyard
This may take a moment while the Shipyard images are pulled..."
    redis=$(docker -H unix://docker.sock run -i -t -d -p 6379:6379 -name shipyard_redis shipyard/redis)
    router=$(docker -H unix://docker.sock run -i -t -d -p 80 -link shipyard_redis:redis -name shipyard_router shipyard/router)
    lb=$(docker -H unix://docker.sock run -i -t -d -p 80:80 -link shipyard_redis:redis -link shipyard_router:app_router -name shipyard_lb shipyard/lb)
    db=$(docker -H unix://docker.sock run -i -t -d -p 5432 -e DB_PASS=$DB_PASS -name shipyard_db shipyard/db)
    shipyard=$(docker -H unix://docker.sock run -i -t -d -p 8000:8000 -link shipyard_db:db -link shipyard_redis:redis -name shipyard -e ADMIN_PASS=$ADMIN_PASS shipyard/shipyard:$TAG app master-worker)
    echo "Configuring Shipyard.  One moment please..."
    # wait for shipyard to start before registering
    until `curl --output /dev/null --silent --head --fail "http://172.17.42.1:8000"`; do
        sleep 1
    done
    KEY=$(curl -d "username=admin&password=shipyard" --silent http://172.17.42.1:8000/api/login | python -c "import sys,json ; data=json.loads(sys.stdin.read()) ; print(data['api_key'])")
    agent=$(docker -H unix://docker.sock run -i -t -d -v /var/run/docker.sock:/docker.sock -p 4500 -name shipyard_agent shipyard/agent -url http://172.17.42.1:8000 -docker /docker.sock -register)
    # wait until host is registered
    until `curl --max-time 5 --output /dev/null --silent --fail -H "Accept: application/json" -H "Authorization: ApiKey admin:$KEY" http://172.17.42.1:8000/api/v1/hosts/1/`; do
        sleep 1
    done
    # set host to enabled
    curl --silent -H "Content-type: application/json" -H "Authorization: ApiKey admin:$KEY" -XPATCH --data '{ "enabled": true}' http://172.17.42.1:8000/api/v1/hosts/1/
    echo "
Shipyard Stack Deployed

You should be able to login with admin:$ADMIN_PASS at http://<docker-host-ip>:8000
You will also need to setup and register the Shipyard Agent.  See http://github.com/shipyard/shipyard-agent for details.
"
elif [ "$ACTION" = "cleanup" ] ; then
    cleanup
    echo "Shipyard Stack Removed"
elif [ "$ACTION" = "purge" ] ; then
    echo "Removing Shipyard images.  This could take a moment..."
    purge
fi


