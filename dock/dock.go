package main

import(
    "fmt"
    "github.com/shrsubra/go-dockerclient"
)

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
    for _,container := range containers{
        fmt.Printf("ID:%s\n", container.ID)
        fmt.Printf("%+v\n", container)
    }
}