# JIRA To ES 

把 jira 中的issue 整体信息全部同步到 ES 中，利用 ES 的文本强大的文本搜索功能来辅助搜索Issue

## 部署前准备工作

1. 安装 ES 
2. 安装 ES 的中文分词插件 IK, [参考文档](https://blog.csdn.net/u011499747/article/details/78917718)
3. 在 google 账号里申请并配置 oauth, [参考文档](https://blog.csdn.net/wangshubo1989/article/details/77980316).通过这一步我们获取到了程序需要的三个相关的参数 google-oauth-client-id, google-oauth-client-secret 和 google-oauth-callback-url, 注意callback的
   URL 为域名(如bug.pingcap.net)加 /auth/callback，其中/auth/callback是代码中的配置。
4. 准备一个可以访问 JIRA 的账号， 包含用户名密码和JIRA URL

## 部署
1. 程序利用static fs 把UI资源编译进了二进制中，如果改了 templates 目录里的 UI 代码， 可以通过执行下面的命令生成新的statik.go文件
`bash
statik -src=templates 
`
2. 编译
`bash
go build main.go
`
3. 执行命令， 注意可以在main.go中查找所以支持的配置。
`bash
./main --jira-password xxxx --jira-username xxx  --es-url http://127.0.0.1:9200
`

## 工作原理
1. 通过 JIRA 的 client 作用 JQL 搜索出所有需要的 ISSUE, JQL是程序的一个配置参数。
2. 定时或者 API 触发后把每个 ISSUE 整体都 存入到 ES 中
3. 把UI请求的关键字转化成下面的 ES 查询语法结构，把高亮结果返回给 UI
```json
{
    "query": {
        "function_score": {
            "query": {
                "bool": {
                    "should": [{
                        "multi_match": %s
                    }]
                }
            }
        }
    },
  
   "highlight": {
    "fields": {
      "*": {}
    }
  }
}
```
