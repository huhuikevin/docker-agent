#!/bin/sh

#放到云主机的clould-init 脚本中运行

#script=/data/server/appAgent/start.sh
#config=/data/server/appAgent/config.yaml
#systempath=/lib/systemd/system
config=config111.yaml
script=start.sh
systempath=./
service=agent.service
servicefile=$systempath/$service
networkdev=eth0
port=8001
proxyserver="http://192.168.10.114:8000"
registry="registry-vpc.cn-hongkong.aliyuncs.com"
username="kevin@1734249857609980"
password="Reg&0928"
image="$registry/lovanow_beta/appagent:v1"
ip=`/sbin/ifconfig $networkdev|grep inet|grep -v inet6|awk '{print $2}'|tr -d "addr:"`
runningsrv="oauth,common-msg,common-file"
create_config()
{
	echo "port: $port" > $config
	echo "proxy:" >> $config
	echo "  server: $proxyserver" >> $config
	echo "  register: /api/v1/register" >> $config
	echo "  keepalive: /api/v1/keepalive" >> $config
	echo "  services: [$runningsrv]" >> $config
	echo "  beatheart: 2" >> $config
	echo "docker:" >> $config
	echo "  cloud: ali" >> $config
	echo "  regin: hongkong" >> $config
	echo "  reposity: $registry" >> $config
	echo "  username: '$username'" >> $config
	echo "  password: '$password'" >> $config
	echo "logs:" >> $config
	echo "  path: /data/server/agent" >> $config
}


create_service()
{
	
	echo "[Unit]" > $servicefile
	echo "Description=Jwaoo Docker agent service" >> $servicefile
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
    echo "	docker run -d --rm --name \$docker_name --net=bridge -v /var/run/docker.sock:/var/run/docker.sock --env APP_PORT=$port --env HOST_IP=$ip -p $port:8000 \$image" >> $script
	echo "}" >> $script

	echo "kill_docker" >> $script
	echo "if [ \"\$mode\"x = \"\"x ];then">> $script
    echo "	start_docker" >> $script
	echo "fi" >> $script

	chmod a+x $script
}

#apt-get update
#apt-get install docker.io

create_config
create_script
create_service

#systemctl daemon-reload

#systemctl enable $service

#systemctl start $service


