# Monit Server

Use godep to cache go dependency.

So if you change or import other go file, run:

```
go build 
godep save
```

Build docker image run

```
docker build --rm -t monitserver .
```


Export docker image 

```
docker save monitserver > monitserver.tar
```

Load your image

```
sudo docker load < monitserver.tar
```


Run docker image in your kuber master node, use host network

```
sudo docker run --privileged=true --restart=on-failure:10 --net=host  -d -v '/etc/ssl/certs:/etc/ssl/certs' monitserver
```

Remember to open your 50000 port.