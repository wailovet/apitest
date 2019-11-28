# apitest
### 简单粗暴的通过测试接口结果生成徽章程序

## 使用
### 利用Yapi导出配置生成第一份文件
> apitest.exe -yapi={Yapi导出配置文件}  -js=apitest.js

#### 生成徽章
把上面的apitest.js放到服务器上,启动徽章模式,访问启用徽章模式的服务器
> apitest.exe -badge -p=80

访问 {服务器地址}/badge.svg?url={apitest.js地址}