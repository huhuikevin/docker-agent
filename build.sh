#!/bin/sh
app=$1

export GOPATH=/Users/huhui/works/source/jwaoo/goProject/go

cd $GOPATH

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -v github.com/huhuikevin/docker-agent/app/$app
go install -v github.com/huhuikevin/docker-agent/app/$app

if [ $? != 0 ];then
	exit 1
fi

dockerfile=Dockerfile
start=start.sh
createDockerfile()
{
	echo "FROM alpine:latest as prod" > $dockerfile
	echo "MAINTAINER kevin.hu <kevin.hu@jwaoo.com>" >> $dockerfile

	echo "RUN apk add --no-cache bash libc6-compat && rm -rf /var/cache/apk/*" >> $dockerfile

	echo "WORKDIR /usr/bin" >> $dockerfile

	echo "COPY $app /usr/bin/" >> $dockerfile

	echo "COPY $start /usr/bin/" >> $dockerfile

	echo "RUN mkdir -p /etc/jwaoo && chmod a+x $app && chmod a+x $start" >> $dockerfile

	echo "CMD [\"$start\"]" >> $dockerfile

}


if [ "$app"x == "proxy"x -o "$app"x == "agent"x ];then
	#exit 0
	rm -rf dockerbuild
	mkdir dockerbuild
	cp $GOPATH/bin/linux_amd64/$app dockerbuild/
	cp $GOPATH/src/github.com/huhuikevin/docker-agent/start.sh dockerbuild/
	cd dockerbuild/

	createDockerfile

	image=`echo $app | tr 'A-Z' 'a-z'`
        image=docker-$image
	echo $image
	docker build -t $image:v1 .
	cd -
	rm -rf dockerbuild
	docker tag $image:v1 192.168.10.250:1180/sensenow/$image:v1
	docker push 192.168.10.250:1180/sensenow/$image:v1

	docker tag $image:v1  registry.cn-hongkong.aliyuncs.com/lovanow_beta/$image:v1
	docker push registry.cn-hongkong.aliyuncs.com/lovanow_beta/$image:v1
fi
#cp go/bin/linux_amd64/beatmanager go/docker/
#cp go/src/jwaoo.com/beatmanager/config_ali.yml go/docker/config.yml

#cd go/docker

#docker build -t logproducer .

#docker tag logproducer:latest 192.168.10.250:1180/sensenow/logproducer:v1
#docker push 192.168.10.250:1180/sensenow/logproducer:v1

#docker tag logproducer:latest registry.cn-hongkong.aliyuncs.com/lovanow_beta/logproducer:v1
#docker push registry.cn-hongkong.aliyuncs.com/lovanow_beta/logproducer:v1

#cd -

