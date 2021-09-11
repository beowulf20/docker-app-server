package main

import (
	"io/ioutil"
	"log"

	registry "github.com/beowulf20/docker-delta-update-server/framework/app-registry"
	framework_rest "github.com/beowulf20/docker-delta-update-server/framework/rest"
	"github.com/docker/docker/client"
)

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	reg, err := registry.NewAppRegistry()
	fatalOnError(err)

	composeFile, err := ioutil.ReadFile("/home/master/documents/personal/docker-diff-deploy-server/build/docker-compose.yml")
	fatalOnError(err)

	app, err := registry.NewApp("app_test_influx", string(composeFile))
	fatalOnError(err)
	fatalOnError(reg.AddApp(app))

	cli, err := client.NewClientWithOpts()
	fatalOnError(err)

	

	// appReplicas, err := reg.ListApps()
	// fatalOnError(err)

	// for _, appReplica := range appReplicas {
	// 	conts, err := utils.AssociateContainerApp(appReplica, cli)
	// 	fatalOnError(err)
	// 	for _, cont := range conts {
	// 		if cont.Container == nil {
	// 			// create container
	// 			cli.ContainerCreate(context.Background(), &container.Config{
	// 				Image: cont.Service.Image,
	// 			}, nil, nil, nil, appReplica.Name+"_"+cont.Service.Name)
	// 		} else {
	// 			cli.ContainerStart(context.Background(), cont.Container.ID, types.ContainerStartOptions{})
	// 		}

	// 	}
	// }

	fatalOnError(framework_rest.NewRestServer(reg, cli))

	// time.Sleep(time.Second * 2)

	// var errs []error
	// for _, appReplica := range appReplicas {
	// 	err = StopContainers(context.Background(), appReplica, cli)
	// 	if err != nil {
	// 		errs = append(errs, err)
	// 	}
	// }

	// if len(errs) > 0 {
	// 	log.Fatal(errs)
	// }

}
