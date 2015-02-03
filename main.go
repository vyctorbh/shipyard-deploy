package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/samalba/dockerclient"
)

var (
	docker dockerclient.Client
)

func before(c *cli.Context) error {
	d, err := dockerclient.NewDockerClient(c.GlobalString("docker"), nil)
	if err != nil {
		fmt.Printf("unable to connect to docker at %s: %s\n", c.GlobalString("docker"), err)
		os.Exit(1)
	}
	docker = d
	return nil
}

func startShipyard(tag, port string) error {
	shipyardImage := fmt.Sprintf("shipyard/shipyard:%s", tag)
	fmt.Printf("Pulling image: %s\n", shipyardImage)
	if err := docker.PullImage(shipyardImage, nil); err != nil {
		fmt.Printf("Error pulling image %s: %s\n", shipyardImage, err)
		os.Exit(1)
	}
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
	rethinkdbImage := "shipyard/rethinkdb"
	shipyardRethinkdbDataConfig := &dockerclient.ContainerConfig{
		Image:      rethinkdbImage,
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
		Image: rethinkdbImage,
		HostConfig: dockerclient.HostConfig{
			PublishAllPorts: true,
			VolumesFrom:     []string{"shipyard-rethinkdb-data"},
			RestartPolicy: dockerclient.RestartPolicy{
				Name:              "always",
				MaximumRetryCount: 0,
			},
		},
	}

	// pull images
	fmt.Printf("Pulling image: %s\n", rethinkdbImage)
	if err := docker.PullImage(rethinkdbImage, nil); err != nil {
		fmt.Printf("error pulling shipyard image: %s\n", err)
		os.Exit(1)
	}

	// start the show
	fmt.Println("Starting Rethinkdb Data")
	rethinkdbDataId, err := docker.CreateContainer(shipyardRethinkdbDataConfig, "shipyard-rethinkdb-data")
	if err != nil {
		fmt.Printf("error creating shipyard rethinkdb data container: %s\n", err)
		os.Exit(1)
	}
	if err := docker.StartContainer(rethinkdbDataId, &shipyardRethinkdbDataConfig.HostConfig); err != nil {
		fmt.Printf("error starting shipyard rethinkdb data container: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting Rethinkdb")
	rethinkdbId, err := docker.CreateContainer(shipyardRethinkdbConfig, "shipyard-rethinkdb")
	if err != nil {
		fmt.Printf("error creating shipyard rethinkdb container: %s\n", err)
		os.Exit(1)
	}
	if err := docker.StartContainer(rethinkdbId, &shipyardRethinkdbConfig.HostConfig); err != nil {
		fmt.Printf("error starting shipyard rethinkdb container: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting Shipyard")
	if err := startShipyard(c.GlobalString("tag"), c.GlobalString("port")); err != nil {
		fmt.Printf("error starting shipyard: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Shipyard Stack started successfully")
	fmt.Println(" Username: admin Password: shipyard")
}

func stopAction(c *cli.Context) {
	fmt.Println("Stopping Shipyard")
	if err := docker.StopContainer("shipyard", 5); err != nil {
		fmt.Printf("error stopping shipyard: %s\n", err)
	}

	fmt.Println("Stopping Shipyard Rethinkdb")
	if err := docker.StopContainer("shipyard-rethinkdb", 5); err != nil {
		fmt.Printf("error stopping shipyard-rethinkdb: %s\n", err)
	}

	fmt.Println("Stopping Shipyard Rethinkdb Data")
	if err := docker.StopContainer("shipyard-rethinkdb-data", 5); err != nil {
		fmt.Printf("error stopping shipyard-rethinkdb-data: %s\n", err)
	}
}

func restartAction(c *cli.Context) {
	fmt.Println("Restarting Shipyard Rethinkdb Data")
	if err := docker.RestartContainer("shipyard-rethinkdb-data", 5); err != nil {
		fmt.Printf("error restarting shipyard-rethinkdb-data: %s\n", err)
	}

	fmt.Println("Restarting Shipyard Rethinkdb")
	if err := docker.RestartContainer("shipyard-rethinkdb", 5); err != nil {
		fmt.Printf("error restarting shipyard-rethinkdb: %s\n", err)
	}

	fmt.Println("Restarting Shipyard")
	if err := docker.RestartContainer("shipyard", 5); err != nil {
		fmt.Printf("error restarting shipyard: %s\n", err)
	}
}

func upgradeAction(c *cli.Context) {
	tag := c.GlobalString("tag")
	fmt.Printf("Upgrading Shipyard to version: %s\n", tag)

	shipyardImage := fmt.Sprintf("shipyard/shipyard:%s", tag)
	if err := docker.PullImage(shipyardImage, nil); err != nil {
		fmt.Printf("error pulling shipyard image: %s\n", err)
		os.Exit(1)
	}

	if err := docker.RemoveContainer("shipyard", true); err != nil {
		fmt.Printf("error removing shipyard: %s\n", err)
	}

	if err := startShipyard(c.GlobalString("tag"), c.GlobalString("port")); err != nil {
		fmt.Printf("error starting shipyard: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Shipyard Upgraded Successfully")
}

func removeAction(c *cli.Context) {
	fmt.Println("Removing Shipyard Rethinkdb Data")
	if err := docker.RemoveContainer("shipyard-rethinkdb-data", true); err != nil {
		fmt.Printf("error removing shipyard-rethinkdb-data: %s\n", err)
	}

	fmt.Println("Removing Shipyard Rethinkdb")
	if err := docker.RemoveContainer("shipyard-rethinkdb", true); err != nil {
		fmt.Printf("error removing shipyard-rethinkdb: %s\n", err)
	}

	fmt.Println("Removing Shipyard")
	if err := docker.RemoveContainer("shipyard", true); err != nil {
		fmt.Printf("error removing shipyard: %s\n", err)
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
