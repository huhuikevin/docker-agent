package uitls

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	client "github.com/huhuikevin/docker-agent/dockerclient"
	log "github.com/huhuikevin/docker-agent/logger"
)

var stopTimeout uint = 10 //10s

//GetAwsToken get aws username&password
func GetAwsToken(regin string) []string {
	token := client.GetECRToken(regin)
	if token == "" {
		return nil
	}
	return strings.Split(token, ":")
}

//PullDockerImage will pull the image from repo
func PullDockerImage(image string, user string, pwd string) error {
	client, err := client.GetClient()
	if err != nil {
		log.Println("get Client error=", err)
		return err
	}

	if user == "" || pwd == "" {
		if err := client.PullImageWithOutAuth(image); err != nil {
			log.Println("pull image error=", err)
			return err
		}
		return nil
	}

	if err := client.PullImageWithAuth(image, user, pwd); err != nil {
		log.Println("pull image error=", err)
		return err
	}
	return nil
}

//StartDocker start the docker
func StartDocker(image string, config client.ContainerConfig, name string) (string, error) {
	client, err := client.GetClient()
	if err != nil {
		log.Println("get Client error=", err)
		return "", err
	}
	id, err := client.StartDocker(image, &config, name)
	log.Println("id = ", id)
	return id, err
}

//FindDockerByName find the running docker by container name
func FindDockerByName(name string) (string, error) {
	client, err := client.GetClient()
	if err != nil {
		log.Println("get Client error=", err)
		return "", err
	}
	id, err := client.ContainerIsRuning(name)
	if err != nil {
		return "", err
	}
	return id, err
}

//FindContainerInfoByName get the container info
func FindContainerInfoByName(name string) map[string]string {
	info := make(map[string]string)
	client, err := client.GetClient()
	if err != nil {
		log.Println("get Client error=", err)
		return info
	}
	return client.FindContainerInfoByName(name)
}

func shellout(command string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	log.Println(err)
}

//StopContainerByName stop the docker by the docker name
func StopContainerByName(name string) error {
	client, err := client.GetClient()
	if err != nil {
		log.Println("get Client error=", err)
		return err
	}
	for {
		id, err := FindDockerByName(name)
		if err != nil {
			return err
		}
		if id != "" {
			err := client.StopContainer(id, stopTimeout)
			if err != nil {
				return err
			}
		} else {
			break
		}
		time.Sleep(time.Duration(2) * time.Second)
	}
	//shellout("docker ps -a")
	return nil
}
