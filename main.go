package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/samalba/dockerclient"
)

var (
	docker dockerclient.Client
)

func before(c *cli.Context) error {
	d, err := dockerclient.NewDockerClient(c.GlobalString("docker"), nil)
	if err != nil {
		log.Fatalf("unable to connect to docker at %s: %s", c.GlobalString("docker"), err)
	}
	docker = d
	return nil
}

func startShipyard(tag, port string) error {
	shipyardImage := fmt.Sprintf("shipyard/shipyard:%s", tag)
	shipyardPort := port
	shipyardConfig := &dockerclient.ContainerConfig{
		Image: shipyardImage,
		HostConfig: dockerclient.HostConfig{
			VolumesFrom: []string{"shipyard-rethinkdb-data"},
			RestartPolicy: dockerclient.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
			PortBindings: map[string][]dockerclient.PortBinding{
				"8080/tcp": []dockerclient.PortBinding{
					{
						HostPort: shipyardPort,
					},
				},
			},
			Links: []string{"shipyard-rethinkdb:rethinkdb"},
		},
	}

	shipyardId, err := docker.CreateContainer(shipyardConfig, "shipyard")
	if err != nil {
		return fmt.Errorf("error creating shipyard container: %s", err)
	}
	if err := docker.StartContainer(shipyardId, &shipyardConfig.HostConfig); err != nil {
		return fmt.Errorf("error starting shipyard container: %s", err)
	}
	return nil
}

func deployAction(c *cli.Context) {
	shipyardRethinkdbDataConfig := &dockerclient.ContainerConfig{
		Image:      "shipyard/rethinkdb",
		Entrypoint: []string{"/bin/bash"},
		Cmd:        []string{"-l"},
		Tty:        true,
		OpenStdin:  true,
		HostConfig: dockerclient.HostConfig{
			RestartPolicy: dockerclient.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
	}
	shipyardRethinkdbConfig := &dockerclient.ContainerConfig{
		Image: "shipyard/rethinkdb",
		HostConfig: dockerclient.HostConfig{
			PublishAllPorts: true,
			VolumesFrom:     []string{"shipyard-rethinkdb-data"},
			RestartPolicy: dockerclient.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
	}

	// start the show
	log.Info("Starting Rethinkdb Data")
	rethinkdbDataId, err := docker.CreateContainer(shipyardRethinkdbDataConfig, "shipyard-rethinkdb-data")
	if err != nil {
		log.Fatalf("error creating shipyard rethinkdb data container: %s", err)
	}
	if err := docker.StartContainer(rethinkdbDataId, &shipyardRethinkdbDataConfig.HostConfig); err != nil {
		log.Fatalf("error starting shipyard rethinkdb data container: %s", err)
	}

	log.Info("Starting Rethinkdb")
	rethinkdbId, err := docker.CreateContainer(shipyardRethinkdbConfig, "shipyard-rethinkdb")
	if err != nil {
		log.Fatalf("error creating shipyard rethinkdb container: %s", err)
	}
	if err := docker.StartContainer(rethinkdbId, &shipyardRethinkdbConfig.HostConfig); err != nil {
		log.Fatalf("error starting shipyard rethinkdb container: %s", err)
	}

	log.Info("Starting Shipyard")
	if err := startShipyard(c.GlobalString("tag"), c.GlobalString("port")); err != nil {
		log.Fatalf("error starting shipyard: %s", err)
	}

	log.Info("Shipyard Stack started successfully")

}

func stopAction(c *cli.Context) {
	log.Info("Stopping Shipyard")
	if err := docker.StopContainer("shipyard", 5); err != nil {
		log.Errorf("error stopping shipyard: %s", err)
	}

	log.Info("Stopping Shipyard Rethinkdb")
	if err := docker.StopContainer("shipyard-rethinkdb", 5); err != nil {
		log.Errorf("error stopping shipyard-rethinkdb: %s", err)
	}

	log.Info("Stopping Shipyard Rethinkdb Data")
	if err := docker.StopContainer("shipyard-rethinkdb-data", 5); err != nil {
		log.Errorf("error stopping shipyard-rethinkdb-data: %s", err)
	}
}

func restartAction(c *cli.Context) {
	log.Info("Restarting Shipyard Rethinkdb Data")
	if err := docker.RestartContainer("shipyard-rethinkdb-data", 5); err != nil {
		log.Errorf("error restarting shipyard-rethinkdb-data: %s", err)
	}

	log.Info("Restarting Shipyard Rethinkdb")
	if err := docker.RestartContainer("shipyard-rethinkdb", 5); err != nil {
		log.Errorf("error restarting shipyard-rethinkdb: %s", err)
	}

	log.Info("Restarting Shipyard")
	if err := docker.RestartContainer("shipyard", 5); err != nil {
		log.Errorf("error restarting shipyard: %s", err)
	}
}

func upgradeAction(c *cli.Context) {
	tag := c.GlobalString("tag")
	log.Infof("Upgrading Shipyard to version: %s", tag)

	shipyardImage := fmt.Sprintf("shipyard/shipyard:%s", tag)
	if err := docker.PullImage(shipyardImage, nil); err != nil {
		log.Fatalf("error pulling shipyard image: %s", err)
	}

	if err := docker.RemoveContainer("shipyard", true); err != nil {
		log.Warnf("error removing shipyard: %s", err)
	}

	if err := startShipyard(c.GlobalString("tag"), c.GlobalString("port")); err != nil {
		log.Fatalf("error starting shipyard: %s", err)
	}

	log.Info("Shipyard Upgraded Successfully")
}

func removeAction(c *cli.Context) {
	log.Info("Removing Shipyard Rethinkdb Data")
	if err := docker.RemoveContainer("shipyard-rethinkdb-data", true); err != nil {
		log.Errorf("error removing shipyard-rethinkdb-data: %s", err)
	}

	log.Info("Removing Shipyard Rethinkdb")
	if err := docker.RemoveContainer("shipyard-rethinkdb", true); err != nil {
		log.Errorf("error removing shipyard-rethinkdb: %s", err)
	}

	log.Info("Removing Shipyard")
	if err := docker.RemoveContainer("shipyard", true); err != nil {
		log.Errorf("error removing shipyard: %s", err)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "shipyard-deploy"
	app.Usage = "deploy a shipyard stack"
	app.Author = "shipyard project"
	app.Before = before
	app.Email = "shipyard-project.com"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "deploy stack",
			Action: deployAction,
		},
		{
			Name:   "stop",
			Usage:  "stop shipyard stack",
			Action: stopAction,
		},
		{
			Name:   "restart",
			Usage:  "restart shipyard stack",
			Action: restartAction,
		},
		{
			Name:   "upgrade",
			Usage:  "upgrade shipyard stack",
			Action: upgradeAction,
		},
		{
			Name:   "remove",
			Usage:  "remove shipyard stack",
			Action: removeAction,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "docker, d",
			Value: "unix:///var/run/docker.sock",
			Usage: "url to Docker",
		},
		cli.StringFlag{
			Name:  "tag, t",
			Value: "latest",
			Usage: "tag of shipyard to deploy",
		},
		cli.StringFlag{
			Name:  "port, p",
			Value: "8080",
			Usage: "port to run shipyard controller",
		},
	}
	app.Run(os.Args)
}
