# NOTE: this is no longer used in Shipyard v2. This repository is left here for legacy purposes.

See http://shipyard-project.com for v2 information.

# Shipyard Deploy
This will deploy a componentized Shipyard stack on your Docker host.  It will
pull, launch, and link the various services together so you have an entire
Shipyard stack.

# Usage
You must bind the Docker socket into the container in order for the deploy container
to work with the Docker host.

# Environment Variables
There are a few environment variables to allow you to customize the stack:

* `DB_PASS`: set the Shipyard Database instance user password
* `ADMIN_PASS`: set the Shipyard admin account password (default: shipyard)
* `TAG`: The tag to use for the Shipyard instance (default: latest)
* `DB_HOST_VOLUME`: (optional) Specify a host directory to map the Shipyard DB volume
* `DEBUG`: (optional) Enable debug in Shipyard (default: false)

## Setup Shipyard Stack
`docker run -i -t -v /var/run/docker.sock:/docker.sock shipyard/deploy setup`

After running setup you will need to install the agent.  The easiest way is to use the Docker image as mentioned in the Agent docs at https://github.com/shipyard/shipyard-agent#docker

## Remove Shipyard Stack
`docker run -i -t -v /var/run/docker.sock:/docker.sock shipyard/deploy cleanup`

## Upgrade Shipyard
`docker run -i -t -v /var/run/docker.sock:/docker.sock shipyard/deploy upgrade`
