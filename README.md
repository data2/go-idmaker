# 安全可靠的高性能分布式序列自增基服务

[![License](http://img.shields.io/:license-apache-brightgreen.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)
[![Build Status](https://travis-ci.org/data2/go-idmaker.svg?branch=master)](https://travis-ci.org/data2/go-idmaker)
 
安全可靠的高性能分布式序列自增基服务go-idmaker

多个app访问基服务，获取到基id，默认获取[id*1000,(id+1)*1000)序列，使用完毕后再请求获取新基id，类似于java中的分段锁，整体架构并发安全，性能最高

# architecture

![go-idmaker (2)](https://user-images.githubusercontent.com/13504729/131777983-6b274c5e-765e-4e0a-90a9-638e0bb13988.png)

# 启动项目

```go

go run idmaker.go 8080, /idMaker

启动项目后，访问http://ip:${port}/${path}

```

# testing

go test -v -run TestPrettyClientReturn  idmaker_test.go idmaker.go 

