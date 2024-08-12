# ztlib



### 运行所有测试
可以使用以下命令运行项目中的所有测试：

```sh
go test -tags='!ignore' ./...

golangci-lint run --disable-all -E goimports,misspell,whitespace
```