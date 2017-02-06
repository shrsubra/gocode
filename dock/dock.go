package main

import(
    "fmt"
    "github.com/shrsubra/go-dockerclient"
    "sync"
    "log"
)

type Stat struct {
    containerId docker.APIContainers
    stats *docker.Stats
}

func main() {
    endpoint := "unix:///var/run/docker.sock"
    client,err := docker.NewClient(endpoint)
    if err != nil{
        panic(err)
    }
    containers, err := client.ListContainers(docker.ListContainersOptions{})
    if err != nil {
        panic(err)
    }

    // logic:
    // a top level waitgroup that waits for all containers to write Stat to the queue channel
    // each
    var topWg sync.WaitGroup
    topWg.Add(len(containers))

    queue := make(chan Stat, 1)
    for _,container := range containers{
        fmt.Printf("new go func for container:%v\n", container.ID)
        cont, err := client.InspectContainer(container.ID)
        if err == nil{
            fmt.Printf("Inspect container:%+v\n", cont)
        } else {
            log.Fatal(err)
        }
        top, err := client.TopContainer(container.ID, "")
        if err == nil{
            fmt.Printf("Top container:%+v\n", top)
        } else {
            log.Fatal(err)
        }
        go func(container docker.APIContainers){
            queue <- getMetricForContainer(client, &container)
        }(container)
    }
    go func() {
        for contStat := range queue {
            fmt.Printf("containerID: %s, stats: %+v\n", contStat.containerId.ID, contStat.stats)
            topWg.Done()
        }
    }()

    topWg.Wait()
}

func getMetricForContainer(client *docker.Client, container *docker.APIContainers) Stat {
    var event Stat
    var wg sync.WaitGroup
    wg.Add(2)
    statC := make(chan *docker.Stats)
    errC := make(chan error, 1)
    go func() {
        defer wg.Done()
        errC <- client.Stats(docker.StatsOptions{ID: container.ID, Stats: statC})
        close(errC)
    }()
    go func() {
        defer  wg.Done()
        stats := <- statC
        err := <- errC
        if err == nil {
            event.stats = stats
            event.containerId = *container
        }
       // close(statC)
    }()
    wg.Wait()
    return event
}