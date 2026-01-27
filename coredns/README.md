# coredns

``` bash
cd coredns
go get github.com/Eun/coredns-ipecho
go get github.com/wenerme/coredns-ipin
go generate && go build
./coredns -plugins
cd ..
docker build -t tlsu-coredns .
docker run -it --rm tlsu-coredns -plugins
```
