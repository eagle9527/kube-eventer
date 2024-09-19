FROM  registry.aliyuncs.com/acs/kube-eventer:v1.2.7-ca03be0-aliyun
ADD kube-eventer /
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build -o  kube-eventer  .
#docker build  --platform=amd64  -t  kube-eventer:v1.2.7-ca03be2-aliyun  .
#docker save   kube-eventer:v1.2.7-ca03be2-aliyun > kube-eventer-v1.2.7.tar
ENTRYPOINT ["/kube-eventer"]

