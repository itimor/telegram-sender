# telegram-sender

Nightingale的理念，是将告警事件扔到redis里就不管了，接下来由各种sender来读取redis里的事件并发送，毕竟发送报警的方式太多了，适配起来比较费劲，希望社区同仁能够共建。

这里提供一个微信的sender，参考了[https://github.com/yanjunhui/chat](https://github.com/yanjunhui/chat)，具体如何获取企业微信信息，也可以参看yanjunhui这个repo

## compile

```bash
cd $GOPATH/src
mkdir -p github.com/itimor
cd github.com/itimor
git clone https://github.com/itimor/telegram-sender.git
cd telegram-sender
# 该项目不需要mod,临时关闭,如其他项目需要记得在env开启来
export GO111MODULE=off
go build
```

如上编译完就可以拿到二进制了。

## configuration

直接修改etc/telegram-sender.yml即可。另外itimor-monapi这个模块默认的发送通道只打开了mail，如果要同时使用im，需要在notify这里打开相关配置：

```yaml
notify:
  p1: ["mail", "im"]
  p2: ["mail", "im"]
  p3: ["mail", "im"]
```

## pack

编译完成之后可以打个包扔到线上去跑，将二进制和配置文件打包即可：

```bash
tar zcvf telegram-sender.tar.gz telegram-sender etc/telegram-sender.yml etc/telegram.tpl
```

## test

配置etc/telegram-sender.yml，相关配置修改好，我们先来测试一下是否好使， `./telegram-sender -t <toUser>`，程序会自动读取etc目录下的配置文件，发一个测试消息给`toUser`

## run

如果测试发送没问题，扔到线上跑吧，使用systemd或者supervisor之类的托管起来，systemd的配置实例：


```
$ cat telegram-sender.service
[Unit]
Description=Nightingale telegram sender
After=network-online.target
Wants=network-online.target

[Service]
User=root
Group=root

Type=simple
ExecStart=/data/app/n9e/telegram-sender
WorkingDirectory=/data/app/n9e

Restart=always
RestartSec=1
StartLimitInterval=0

[Install]
WantedBy=multi-user.target
```
