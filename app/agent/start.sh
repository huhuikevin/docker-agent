#!/bin/sh
mode=$1
registry=registry-vpc.cn-hongkong.aliyuncs.com
image=$registry/lovanow_beta/appagent:v1
docker_name=appagent
kill_docker()
{
	while [ true ];do
		runningid=`docker ps --filter=name=$docker_name --format {{.ID}}`
		if [ "$runningid"x != ""x ];then
			docker kill $runningid
			sleep 2
		else
			break
		fi
	done
}
start_docker()
{
	docker login -u "kevin@1734249857609980" -p "Reg&0928" $registry
	docker run -d --rm --name $docker_name --net=bridge -v /var/run/docker.sock:/var/run/docker.sock --env APP_PORT=8000 -p 8001:8000 $image
}
kill_docker
if [ "$mode"x == ""x ];then
	start_docker
fi
