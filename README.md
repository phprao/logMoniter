# logMoniter

#### 介绍
##### nginx日志实时监控系统。
##### 采用协程模式实现日志的读取，分析，存储。
##### 使用 influxdb + grafana 的配合完美的存储与展现监控的指标。

#### 软件架构
##### access.log 为示例日志文件
##### main.go 为主程序，同时提供restful api 来展示主程序的健康状况，地址为 http://localhost:9193/monitor
##### send.go 为向access.log写入日志的程序，模拟日志的实时写入。

#### 安装教程

1.  启动程序 go run main.go
2.  模拟日志写入 go run send.go
3.  主程序健康状况 http://localhost:9193/monitor
4.  读通道和写通道为0表示没有待处理的数据
5.  tps表示平均每秒处理的记录数
#### 使用说明

1.  xxxx
2.  xxxx
3.  xxxx

#### 参与贡献

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request


#### 码云特技

1.  使用 Readme\_XXX.md 来支持不同的语言，例如 Readme\_en.md, Readme\_zh.md
2.  码云官方博客 [blog.gitee.com](https://blog.gitee.com)
3.  你可以 [https://gitee.com/explore](https://gitee.com/explore) 这个地址来了解码云上的优秀开源项目
4.  [GVP](https://gitee.com/gvp) 全称是码云最有价值开源项目，是码云综合评定出的优秀开源项目
5.  码云官方提供的使用手册 [https://gitee.com/help](https://gitee.com/help)
6.  码云封面人物是一档用来展示码云会员风采的栏目 [https://gitee.com/gitee-stars/](https://gitee.com/gitee-stars/)
