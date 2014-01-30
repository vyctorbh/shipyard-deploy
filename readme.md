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
* `REVISION`: The tag to use for the Shipyard instance (default: latest)

## Setup Shipyard Stack
`docker run -i -t -v /var/run/docker.sock:/docker.sock shipyard/deploy setup`

## Remove Shipyard Stack
`docker run -i -t -v /var/run/docker.sock:/docker.sock shipyard/deploy cleanup`
