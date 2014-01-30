from ubuntu:12.04
maintainer Shipyard Project "http://shipyard-project.com"
run apt-get update
run apt-get install -y curl
run curl https://get.docker.io/builds/Linux/x86_64/docker-latest -o /usr/local/bin/docker
run chmod +x /usr/local/bin/docker
add run.sh /usr/local/bin/run
entrypoint ["/usr/local/bin/run"]
