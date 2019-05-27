package client

import (
	"bytes"
	"errors"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/huhuikevin/docker-agent/logger"
	//"golang.org/x/net/context"
)

//Client docker client wrapper
type Client struct {
	// client used to send and receive http requests.
	client *docker.Client
}

//ULimit linux system limite
type ULimit struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Soft int64  `json:"soft,omitempty" yaml:"soft,omitempty" toml:"soft,omitempty"`
	Hard int64  `json:"hard,omitempty" yaml:"hard,omitempty" toml:"soft,omitempty"`
}

//ContainerConfig docker start config
type ContainerConfig struct {
	Env         []string          `json:"Env,omitempty" yaml:"Env,omitempty" toml:"Env,omitempty"`
	Volume      []string          `json:"Volume,omitempty" yaml:"Volume,omitempty" toml:"Volume,omitempty"`
	Port        map[string]string `json:"Port,omitempty" yaml:"Port,omitempty" toml:"Port,omitempty"`
	Cmd         []string          `json:"Cmd,omitempty" yaml:"Cmd,omitempty" toml:"Cmd,omitempty"`
	NetworkMode string            `json:"NetworkMode,omitempty" yaml:"NetworkMode,omitempty" toml:"NetworkMode,omitempty"`
	Hostname    string            `json:"Hostname,omitempty" yaml:"Hostname,omitempty" toml:"Hostname,omitempty"`
	Memory      int64             `json:"Memory,omitempty" yaml:"Memory,omitempty" toml:"Memory,omitempty"`
	MemorySwap  int64             `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty" toml:"MemorySwap,omitempty"`
	Ulimits     []ULimit          `json:"Ulimits,omitempty" yaml:"Ulimits,omitempty" toml:"Ulimits,omitempty"`
	ExtraHosts  []string          `json:"ExtraHosts,omitempty" yaml:"ExtraHosts,omitempty" toml:"ExtraHosts,omitempty"`
	Labels      map[string]string `json:"Labels,omitempty" yaml:"Labels,omitempty" toml:"Labels,omitempty"`
}

//GetClient get an new client
func GetClient() (*Client, error) {
	cli, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	client := &Client{
		client: cli,
	}
	return client, nil
}

//PullImage pull docker image from register
func (cli *Client) PullImage(imageName string) error {
	if strings.Contains(imageName, "amazonaws") {
		part := strings.Split(imageName, ".")
		if len(part) < 6 {
			return errors.New("not valide")
		}
		regin := part[3]
		token := GetECRToken(regin)
		tokenPart := strings.Split(token, ":")
		return cli.PullImageWithAuth(imageName, tokenPart[0], tokenPart[1])
	}
	return cli.PullImageWithOutAuth(imageName)
}

//PullImageWithOutAuth with no username&password
func (cli *Client) PullImageWithOutAuth(imageName string) error {
	var buf bytes.Buffer
	nameParts := strings.Split(imageName, ":")

	opts := docker.PullImageOptions{
		Repository:   imageName,
		OutputStream: &buf,
	}
	if len(nameParts) == 2 {
		opts.Tag = nameParts[1]
	}
	err := cli.client.PullImage(opts, docker.AuthConfiguration{})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(buf.String())
	return nil
}

//PullImageWithAuth with pass&name
func (cli *Client) PullImageWithAuth(imageName string, user string, pwd string) error {
	var buf bytes.Buffer
	nameParts := strings.Split(imageName, ":")

	opts := docker.PullImageOptions{
		Repository:   imageName,
		OutputStream: &buf,
	}
	if len(nameParts) == 2 {
		opts.Tag = nameParts[1]
	} else {
		opts.Tag = "latest"
	}
	err := cli.client.PullImage(opts, docker.AuthConfiguration{
		Username: user,
		Password: pwd,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(buf.String())
	return nil
}

//StartDocker start the docker
func (cli *Client) StartDocker(imageName string, config *ContainerConfig, containerName string) (string, error) {
	//ctx := context.Background()
	// err := cli.PullImage(imageName)
	// if err != nil {
	// 	log.Println(err)
	// 	return "", err
	// }

	dockerEnv := make([]string, len(config.Env))
	copy(dockerEnv, config.Env)
	// for envKey, envVal := range config.Env {
	// 	dockerEnv = append(dockerEnv, envKey+"="+envVal)
	// }
	portmap := make(map[docker.Port][]docker.PortBinding)
	for portKey, portVal := range config.Port {
		dockerPort := docker.Port(portKey + "/" + "tcp")
		portmap[dockerPort] = []docker.PortBinding{{HostPort: portVal}}
	}
	//log.Println(portmap)
	hostbind := make([]string, len(config.Volume))
	copy(hostbind, config.Volume)
	// var count = 0
	// for volKey, volVal := range config.Volume {
	// 	hostbind[count] = volKey + ":" + volVal
	// 	count++
	// }
	//log.Println(hostbind)

	dockerExposedPorts := make(map[docker.Port]struct{})
	for portKey := range config.Port {
		dockerPort := docker.Port(portKey + "/" + "tcp")
		dockerExposedPorts[dockerPort] = struct{}{}
	}
	dockerconfig := &docker.Config{
		Image:        imageName,
		Env:          dockerEnv,
		Cmd:          config.Cmd,
		ExposedPorts: dockerExposedPorts,
		Tty:          true,
		AttachStdin:  true,
		Hostname:     config.Hostname,
		Labels:       config.Labels,
	}
	ulimits := make([]docker.ULimit, 0, len(config.Ulimits))
	for _, v := range config.Ulimits {
		u := docker.ULimit{Name: v.Name, Soft: v.Soft, Hard: v.Hard}
		ulimits = append(ulimits, u)
	}
	hostconfig := &docker.HostConfig{
		Binds:        hostbind,
		PortBindings: portmap,
		NetworkMode:  config.NetworkMode,
		Ulimits:      ulimits,
		Memory:       config.Memory,
		MemorySwap:   config.MemorySwap,
		ExtraHosts:   config.ExtraHosts,
		AutoRemove:   true,
	}

	container, err := cli.client.CreateContainer(docker.CreateContainerOptions{
		Name:       containerName,
		Config:     dockerconfig,
		HostConfig: hostconfig,
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	if err := cli.client.StartContainer(container.ID, &docker.HostConfig{}); err != nil {
		log.Println(err)
		return "", err
	}
	return container.ID, nil
}

//StopContainer stop docker container
func (cli *Client) StopContainer(id string, timeout uint) error {
	return cli.client.StopContainer(id, timeout)
}

//FindContainerByName get container by it's name
func (cli *Client) FindContainerByName(name string) ([]docker.APIContainers, error) {
	//ctx := context.Background()

	filters := make(map[string][]string)
	filters["status"] = []string{"running"}
	filters["name"] = []string{name}

	opt := docker.ListContainersOptions{
		Filters: filters,
	}
	contains, err := cli.client.ListContainers(opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return contains, err
}

//FindContainerInfoByLabels get container by it's name
func (cli *Client) FindContainerInfoByLabels(name, value string) map[string]string {
	//ctx := context.Background()

	filters := make(map[string][]string)
	filters["status"] = []string{"running"}
	filters["label"] = []string{name + "=" + value}

	opt := docker.ListContainersOptions{
		Filters: filters,
	}
	info := make(map[string]string)
	contains, err := cli.client.ListContainers(opt)
	if err != nil {
		log.Println(err)
		return info
	}
	for _, container := range contains {
		info["Image"] = container.Image
		info["ID"] = container.ID
		info["Name"] = container.Names[0]
	}
	return info
}

//FindContainerInfoByName wrapper for FindContainerByName
func (cli *Client) FindContainerInfoByName(name string) map[string]string {
	info := make(map[string]string)
	contains, err := cli.FindContainerByName(name)
	if err != nil {
		log.Println(err)
		return info
	}
	for _, container := range contains {
		myname := "/" + name
		if myname == container.Names[0] {
			info["Image"] = container.Image
			info["ID"] = container.ID
			info["Name"] = container.Names[0]
		}
	}
	return info
}

//FindContainerByShortID get container by it's short id
func (cli *Client) FindContainerByShortID(id string) ([]docker.APIContainers, error) {
	//ctx := context.Background()

	filters := make(map[string][]string)
	filters["status"] = []string{"running"}
	filters["id"] = []string{id}

	opt := docker.ListContainersOptions{
		Filters: filters,
	}
	contains, err := cli.client.ListContainers(opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return contains, err
}

//ContainerIsRuning return true if contianer is running
func (cli *Client) ContainerIsRuning(name string) (string, error) {
	contains, err := cli.FindContainerByName(name)
	if err != nil {
		log.Println(err)
		return "", err
	}
	for _, container := range contains {
		myname := "/" + name
		if myname == container.Names[0] {
			return container.ID, nil
		}
	}
	return "", nil
}

//ContainerIsRuningByShortID short id
func (cli *Client) ContainerIsRuningByShortID(shortID string) (string, error) {
	contains, err := cli.FindContainerByShortID(shortID)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if len(contains) == 1 {
		return contains[0].ID, nil
	}
	return "", nil
}
