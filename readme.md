# Shipyard Deploy
This will deploy a componentized Shipyard stack on your Docker host.  It will
pull, launch, and link the various services together so you have an entire
Shipyard stack.

# Usage
You must bind the Docker socket into the container in order for the deploy container
to work with the Docker host.

`docker run --rm shipyard/deploy -h`: show help

## Setup Shipyard Stack
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock shipyard/deploy start`

## Stop Shipyard Stack
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock shipyard/deploy stop`

## Restart Shipyard Stack
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock shipyard/deploy restart`

## Upgrade Shipyard
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock shipyard/deploy upgrade`

## Remove Shipyard
`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock shipyard/deploy remove`
