# ChoyKit



游戏服务器基本组件



### 各子包说明


  包        |  描述
------------|------------
pkg/cluster  | 多进程
pkg/codec    | 协议编解码
pkg/log      | 日志API
pkg/protocol | 协议相关
pkg/qnet     | 网络传输
pkg/sched    | 执行器定时器
pkg/uuid     | 分布式id生成


### 包结构组织

参考[goland-standard-project-layout](https://github.com/golang-standards/project-layout)

  包名   |  用途
---------|--------
 cmd      | 应用程序，但是只是通过main函数调用其它包，不包含具体实现
 pkg      | 开放的业务包和库包
 internal | 不开放的业务包和库包
 vendor   | 依赖包

