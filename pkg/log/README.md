# 日志API设计


### 日志API需求

1. 日志分级;
2. 日志存档（文件，ELK等）;


### 日志API实现

1. log包提供一个对外部稳定的API，外部不必关注log包内部是如何实现;
2. 业务package只需要import log即可使用日志，不必在new其它对象；
3. 通过不同的API名称实现日志分级 `log.Infof,log.Warnf, log.Errorf, log.Fatalf`
