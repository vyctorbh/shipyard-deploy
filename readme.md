# Shipyard Deploy
This will deploy a componentized Shipyard stack on your Docker host.  It will
pull, launch, and link the various services together so you have an entire
Shipyard stack.

# Usage
You must bind the Docker socket into the container in order for the deploy container
to work with the Docker host.

## Setup Shipyard Stack
`docker run -v /var/run/docker.sock:/docker.sock shipyard/deploy setup`

## Remove Shipyard Stack
`docker run -v /var/run/docker.sock:/docker.sock shipyard/deploy cleanup`
