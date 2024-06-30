# 基于Token认证的server和client：


```
Token的认证，实际上可以和TLS叠加的，只要PerRPCCredentials接口中第二个方法实现返回true即可。

这里的例子暂时不叠加
```