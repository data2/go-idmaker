# go-idmaker

 安全可靠的高性能分布式序列自增基服务

# 启动项目

```go

go run idmaker.go 8080, /idMaker

启动项目后，访问http://ip:${port}/${path}

```

# testing

go test -v -run TestPrettyClientReturn  idmaker_test.go idmaker.go 

