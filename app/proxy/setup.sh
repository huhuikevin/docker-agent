#!/bin/sh

#放到云主机的clould-init 脚本中运行
#docker registry 地址
registry="registry-vpc.cn-hongkong.aliyuncs.com"
#docker registry 用户名密码
username="sensenow"
password="jwaoo2017$"
#proxy docker 镜像地址
image="$registry/lovanow_beta/docker-proxy:v1"
#proxy 监听端口
port=8080
#proxy 对外暴露的接口，客户端和agent通过这个端口连接
ext_port=2000

mkdir -p /etc/jwaoo
script=/data/server/docker-proxy/start.sh
config=/etc/jwaoo/docker-proxy.yaml
systempath=/lib/systemd/system
#config=config111.yaml
#script=start.sh
#systempath=./
service=docker-proxy.service
servicefile=$systempath/$service


create_config()
{
	echo "port: $port" > $config
	echo "redis:" >> $config
	echo "  server: localhost:6379" >> $config
	echo "  redisDB: 1" >> $config
	echo "checkagent:" >> $config
	echo "  interval: 10" >> $config
	echo "logs:" >> $config
	echo "  path: /logs" >> $config
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
	mkdir -p /data/server/docker-proxy
	echo "#!/bin/sh" > $script

	echo "mode=\$1" >> $script

	echo "registry=$registry" >> $script
	echo "image=$image" >> $script
	echo "docker_name=docker-proxy" >> $script

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
    echo "	docker pull \$image" >> $script
    echo "	docker run -d --rm --name \$docker_name --net=bridge --env APP=proxy -v $config:$config -v /data/server/docker-proxy/logs:/logs -p $ext_port:$port \$image" >> $script
	echo "}" >> $script

	echo "kill_docker" >> $script
	echo "if [ \"\$mode\"x = \"\"x ];then">> $script
    echo "	start_docker" >> $script
	echo "fi" >> $script

	chmod a+x $script
}

apt-get update
apt-get install -y -q docker.io

create_config
create_script
create_service

sleep 5

systemctl daemon-reload

systemctl enable $service

systemctl start $service


