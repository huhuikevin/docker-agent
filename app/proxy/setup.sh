#!/bin/sh

#放到云主机的clould-init 脚本中运行

script=/data/server/appProxy/start.sh
config=/data/server/appProxy/config.yaml
systempath=/lib/systemd/system
#config=config111.yaml
#script=start.sh
#systempath=./
service=proxy.service
servicefile=$systempath/$service
port=8001
proxyserver="http://192.168.10.114:8000"
registry="registry-vpc.cn-hongkong.aliyuncs.com"
username="kevin@1734249857609980"
password="Reg&0928"
image="$registry/lovanow_beta/appproxy:v1"

create_config()
{
	echo "port: $port" > $config
	echo "redis:" >> $config
	echo "  server: localhost:6379" >> $config
	echo "  redisDB: 1" >> $config
	echo "checkagent:" >> $config
	echo "  interval: 10" >> $config
	echo "logs:" >> $config
	echo "  path: /data/server/proxy" >> $config
}


create_service()
{
	
	echo "[Unit]" > $servicefile
	echo "Description=Jwaoo Docker proxy service" >> $servicefile
	echo "After=docker.service" >> $servicefile

	echo "[Service]" >> $servicefile
	echo 'Type=oneshot' >> $servicefile
	echo "ExecStart=$script" >> $servicefile
	echo "ExecStop=$script kill" >> $servicefile
	echo "KillMode=process" >> $servicefile
	echo "RemainAfterExit=yes" >> $servicefile
	#echo "RestartPreventExitStatus=255" >> $servicefile


	echo "[Install]" >> $servicefile
	echo "WantedBy=multi-user.target" >> $servicefile
}


create_script()
{
	mkdir -p /data/server/appAgent
	echo "#!/bin/sh" > $script

	echo "mode=\$1" >> $script

	echo "registry=registry-vpc.cn-hongkong.aliyuncs.com" >> $script
	echo "image=$image" >> $script
	echo "docker_name=appagent" >> $script

	echo "kill_docker()" >> $script
	echo "{" >> $script
    echo "	while [ true ];do" >> $script
    echo "		runningid=\`docker ps --filter="name=\$docker_name" --format "{{.ID}}"\`" >> $script
    echo "		if [ \"\$runningid\"x != \"\"x ];then" >> $script
    echo "			docker kill \$runningid" >> $script
    echo "			sleep 2" >> $script
    echo "		else" >> $script
    echo "			break" >> $script
    echo "		fi" >> $script
    echo "	done" >> $script
	echo "}" >> $script

	echo "start_docker()" >> $script
	echo "{" >> $script
    echo "	docker login -u \"$username\" -p \"$password\" $registry" >> $script
    echo "  docker pull \$image" >> $script
    echo "	docker run -d --rm --name \$docker_name --net=bridge -p $port:8000 \$image" >> $script
	echo "}" >> $script

	echo "kill_docker" >> $script
	echo "if [ \"\$mode\"x = \"\"x ];then">> $script
    echo "	start_docker" >> $script
	echo "fi" >> $script

	chmod a+x $script
}

apt-get update
apt-get install docker.io

create_config
create_script
create_service

systemctl daemon-reload

systemctl enable $service

systemctl start $service


