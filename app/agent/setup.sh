#!/bin/sh

#放到云主机的clould-init 脚本中运行
#这台host上需要运行哪些服务
runningsrv="oauth,common-msg,common-file"
#proxy 代理地址，agent启动的时候需要注册到proxy
proxyserver="http://192.168.10.114:2000"
#docker registry 地址
registry="registry-vpc.cn-hongkong.aliyuncs.com"
#docker agent的镜像名称
image="$registry/lovanow_beta/agent:v1"
#用户名密码
username="sensenow"
password="jwaoo2017$"

#agent的监听端口
port=8080
#对外暴露的端口,注册到proxy的时候用host的ip+ext_port
ext_port=2000

mkdir -p /etc/jwaoo

script=/data/server/agent/start.sh
#script=start.sh
config=/etc/jwaoo/agent.yaml
#config=config111.yaml
systempath=/lib/systemd/system
#systempath=./
service=agent.service
servicefile=$systempath/$service
networkdev=eth0
ip=`/sbin/ifconfig $networkdev|grep inet|grep -v inet6|awk '{print $2}'|tr -d "addr:"`

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
	echo "  reposity: $registry" >> $config
	echo "  username: '$username'" >> $config
	echo "  password: '$password'" >> $config
	echo "logs:" >> $config
	echo "  path: /logs" >> $config
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
	mkdir -p /data/server/agent
	echo "#!/bin/sh" > $script

	echo "mode=\$1" >> $script

	echo "registry=$registry" >> $script
	echo "image=$image" >> $script
	echo "docker_name=agent" >> $script

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
    echo "	docker run -d --rm --name \$docker_name --net=bridge -v $config:$config -v /data/server/agent/logs:/logs -v /var/run/docker.sock:/var/run/docker.sock --env APP=agent --env APP_PORT=$ext_port --env HOST_IP=$ip -p $ext_port:$port \$image" >> $script
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

systemctl daemon-reload

systemctl enable $service

systemctl start $service


