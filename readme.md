# 腾讯云监控 Exporter

## 安装

### 环境
`go1.9.x` (and later)
### 编译
```shell
cd $GOPATH/src
go get github.com/tencentyun/tencentcloud-exporter
go build
```

## 快速开始

需要配置腾讯云提供的access_key,secret_key和地区id，例子(qcloud.yml)如下：

```yaml
credential:
  access_key: "your_fancy_accesskey"
  secret_key: "your_fancy_accesskey"
  region: "ap-shanghai"
metrics:
 - tc_namespace: test/cvm
   tc_metric_name: CPUUsage  
   tc_metric_rename: cpu_use  
   tc_labels: [InstanceName,Zone]  
   tc_statistics: [max] 
   period_seconds: 60
   delay_seconds: 300
   range_seconds: 120
```

启动 Exporter

```bash
> tencentcloud-exporter --web.listen-address "127.0.0.1:9123" --config.file "qcloud.yml" --web.telemetry-path "/metrics"
```

访问 [http://127.0.0.1:9123/metrics](http://127.0.0.1:9123/metrics) 查看指标抓取是否成功



## 高级配置
```
credential:
  access_key: <YOUR_ACCESS_KEY>
  access_secret: <YOUR_ACCESS_SECRET>
  region: <REGION>

rate_limit: 10 #限制此实例接口调用频率, 一秒钟可以有几次监控接口调用

metrics:
 - tc_namespace: xxx/CVM #命名空间(xxx是a-z随意定的名字, 而后面的cvm是固定的,是每个产品的名字)
   tc_metric_name: CPUUsage #腾讯云上指标名字
   tc_metric_rename: cpu_usage #上报的指标名字(默认是tc_metric_name)
   tc_myself_dimensions:#使用者自己指定上报维度,一般用不到设置
      appid: 123456789
	  bucket :"test"
   tc_labels: [Zone,InstanceName,ProjectId,Vip,UniqVpcId] #tag, labels会按照tc_labels元素字节序进行排序
   tc_filters: #过滤实例
     InstanceName: test  #InstanceName必须存在test和my才会上报
	 VpcId: vpc-dk8zmwuf #VpcId必须为vpc-dk8zmwuf才会上报
   tc_statistics: [Max]#计算方法 支持Max Min 和Avg (字母大小写无关)
   period_seconds: 60 #数据统计周期
   range_seconds: 300 #取多少数据进行统计(如例子是取300/60+1==6个进行max,min,avg或sum,越多数据越平稳)
   delay_seconds: 600 #数据延时时长
```


上边说明上报例子:

```
xxx_cvm_cpu_usage_max{instance_name="my_a_test",zone="ap-guangzhou-3",project_id:"0",vpc_id="vpc-dk8zmwuf"} 42.0
xxx_cvm_cpu_usage_max{instance_name="my_b_test",zone="ap-guangzhou-1",project_id:"0",vpc_id="vpc-dk8zmwuf"} 7.0
```



## 使用注意
### 1. tc_myself_dimensions风格 vs tc_labels风格
配置存在两种风格,两种风格在字段上不兼容, tc_myself_dimensions风格里边有严格字段要求(大小写也有要求), tc_myself_dimensions和 [tc_filters、tc_labels]是冲突的.各个产品会按照业务模型、api不同选择tc_myself_dimensions风格或者tc_labels风格, 可以参考各产品文档了解产品使用那种风格

- tc_myself_dimensions风格在生成prometheus tags时会按照key的字节序排序


- tc_labels风格在生成prometheus tags时会按照数组元素的字节序排序

如:CDN业务,只能支持tc_myself_dimensions, 里边必须设置"projectId"和"domain"两个属性,多设置属性、少设置或者设置"ProjectId"等请求皆无法拉取到监控数据

### 2.各属性在细微的差异
同一个属性(label)产品api的支持情形不同, 如cvm的InstanceName不支持模糊匹配, 而mysql\redis等产品支持模糊匹配 ,这个可以参考各个产品拉取实例列表的接口

###3.各个指标period_seconds是不同的
如COS的InternetTraffic指标支持60和300,而StdStorage指标是小时级别的,这部分差异比较多需要参考监控的官方文档


##各个产品风格及说明

- **数据库:mysql**   (tc_labels风格)



支持属性:
```
{"Zone","VpcId","SubnetId","InstanceName","InstanceId","ProjectId","Qps",
"EngineVersion",,"RenewFlag","SubnetId","CPU","Memory","Volume","Vip","Vport","CreateTime"}
```


eg:
```
 - tc_namespace: guauga/Mysql
   tc_metric_name: BytesSent  
   tc_metric_rename: MyNewName  
   tc_labels: [ProjectId,Zone]  
   tc_statistics: [Max,Min,Avg] 
   period_seconds: 60
   delay_seconds: 300
   range_seconds: 120
```
- **虚拟主机:cvm**   (tc_labels风格)

支持属性:
```
{"Zone", "VpcId", "SubnetId","InstanceName", "InstanceId", "PrivateIpAddress","PublicIpAddress",
 "InstanceChargeType","InstanceType","CreatedTime","ImageId","RenewFlag","SubnetId","CPU","Memory"}
```

```
 - tc_namespace: guauga/cvm
   tc_metric_name: CPUUsage  
   tc_metric_rename: cpu_use  
   tc_labels: [Zone,InstanceId,InstanceName]
   tc_filters: 
     InstanceName: "dev"
     Zone: "ap-guangzhou-4" 
   tc_statistics: [Max,Min,Avg] 
   period_seconds: 60
   delay_seconds: 300
   range_seconds: 120
```
- **键值存储:redis**   (tc_labels风格)

支持属性:
```
{"Zone", "VpcId", "SubnetId","InstanceName", "InstanceId", "PrivateIpAddress","PublicIpAddress",
 "InstanceChargeType","InstanceType","CreatedTime","ImageId","RenewFlag","SubnetId","CPU","Memory"}
```

```
 - tc_namespace: guauga/cvm
   tc_metric_name: CPUUsage  
   tc_metric_rename: cpu_use  
   tc_labels: [Zone,InstanceId,InstanceName]
   tc_filters: 
     InstanceName: "dev"
     Zone: "ap-guangzhou-4" 
   tc_statistics: [Max,Min,Avg] 
   period_seconds: 60
   delay_seconds: 300
   range_seconds: 120
```
- **负载均衡(公网):public_clb**   (tc_labels风格)

支持属性:
```
{"LoadBalancerName","LoadBalancerVip","ProjectId"}
```

```
 - tc_namespace: Tencent/public_clb
   tc_metric_name: Outtraffic  
   tc_labels: [LoadBalancerVip,ProjectId]
   tc_filters: 
     LoadBalancerName: "SK1"   
   tc_statistics: [Max] 
   period_seconds: 60
   delay_seconds: 120
   range_seconds: 120
```

- **内容分发网络:cdn**   (tc_myself_dimensions风格)
 可用维度 [projectId,domain]
```
 - tc_namespace: guauga/cdn
   tc_metric_name: Requests  
   tc_myself_dimensions:
     projectId: 0 
     domain: "s5.hy.qcloudcdn.com" 
   tc_statistics: [Max] 
   period_seconds: 60
   delay_seconds: 600
   range_seconds: 60
```

- **对象存储:cos**   (tc_myself_dimensions风格)
 可用维度 [appid,bucket]
```
 - tc_namespace: guauga/cos
   tc_metric_name: StdWriteRequests  
   tc_myself_dimensions:
     appid: 1251337138 
     bucket: "test-1251337138" 
   tc_statistics: [Max] 
   period_seconds: 60
   delay_seconds: 300
   range_seconds: 60
```






















