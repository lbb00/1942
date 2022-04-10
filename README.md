# 1942 - 叮咚买菜辅助

Author: 无名氏

## 测试效果

![](./docs/console.png)

## Useage

1. 打开微信叮咚买菜，登录后，抓取接口请求，并在根目录下新建一个 yml 文件配置如下内容：

```yml
# 这些数据从请求header头中获取
cookie: '' # cookie
uid: '' # ddmc-uid
deviceId: '' # ddmc-device-id
longitude: '' # ddmc-longitude 可以不配置
latitude: '' # ddmc-longitude 可以不配置

# 这个是个手机通知,需要的话去app store下载一个
barkId: '' # 可以不配置，但你需要盯紧命令行
```

> 如需运行多个账号，需要配置多个 yml 文件，之所以不放在一个 yml 里是希望使用者启用多个 terminal，方便看到不同账号的抢购结果。

2. 在叮咚 app 或小程序中把要采购的商品加购物车，由于访问量巨大，可能要点好几次才能加成功(叮咚现在每天两场抢购，早上 6:00，8:30，提前半小时爬起来加购物车)

3. 运行 `go run main.go -conf 你的yml名字，如a1对应a1.yml` 测试一下是否可以选择地址

4. 配置 main.go 中的 sleep 时间，开抢前 1 分钟运行本脚本

为了避免被风控，脚本内置的时间 sleep 比较长，推荐开抢前修改下面几个配置为。

WARNING: 该配置不适合在非抢购高峰期运行，若向长时间挂机捡漏，请将休眠时间调高！！！

```go
// main.go

var (
	cartSleepTime    = time.Millisecond * 100
	orderSleepTime   = time.Millisecond * 100
	reserveSleepTime = time.Millisecond * 100
)
```
