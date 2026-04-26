
这里面放的都是已经制定好方案的，但是待完善


1 rpc 为保持git的兼容性，go-wind-admin-rpc软连接到了gow的api，单独进行发布，之后可以引入这个进行适配
2 rpc 目前是采用http-client,但是缺乏token,目前的方案是固定api-key进行调用，即提前申请大模型api-key的东西，然后定时任务这些去调用