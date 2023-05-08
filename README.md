# 腾讯云监控 Exporter v2
## 公告
尊敬的用户，由于资源和人力等客观原因，本插件自2023年05月01日起将不再更新迭代，建议使用腾讯云托管的Prometheus监控服务，可通过【集成中心】-【云监控】功能一键安装采集腾讯云产品基础监控数据。   
产品入口：https://console.cloud.tencent.com/monitor/prometheus   
使用文档指引：https://cloud.tencent.com/document/product/1416/76388   

腾讯云监控已于2022年09月01日开始对超出免费额度的 API 接口的请求进行计费，需要手动开通 API 付费，详见计费说明与开通指引。     
开通页面：https://buy.cloud.tencent.com/APIRequestBuy    
资源消耗页：https://console.cloud.tencent.com/monitor/consumer/products   
计费文档：https://cloud.tencent.com/document/product/248/77914   

通过qcloud exporter将云监控支持的产品监控指标自动批量导出  
(`兼容v1版本`)

## 一、支持的产品列表

产品     | 命名空间 |支持的指标|
--------|---------|----------
数据库MongoDB |QCE/CMONGO|[指标详情](https://cloud.tencent.com/document/product/248/45104)
数据库MySQL(CDB)|QCE/CDB|[指标详情](https://cloud.tencent.com/document/product/248/45147)
Redis标准版|QCE/REDIS(即将下线，不推荐)|[指标详情](https://cloud.tencent.com/document/product/248/45111)
Redis集群版|QCE/REDIS_CLUSTER(即将下线，不推荐)|[指标详情](https://cloud.tencent.com/document/product/248/45111)
数据库Redis(内存版)|QCE/REDIS_MEM|[指标详情](https://cloud.tencent.com/document/product/248/49729)
云服务器CVM|QCE/CVM|[指标详情](https://cloud.tencent.com/document/product/248/6843)
COS|QCE/COS|[指标详情](https://cloud.tencent.com/document/product/248/45140)
CDN|QCE/CDN|[指标详情](https://cloud.tencent.com/document/product/248/45138)
负载均衡CLB(公网)|QCE/LB_PUBLIC|[指标详情](https://cloud.tencent.com/document/product/248/51898)
负载均衡CLB(7层)|QCE/LOADBALANCE|[指标详情](https://cloud.tencent.com/document/product/248/45045)
NAT网关|QCE/NAT_GATEWAY|[指标详情](https://cloud.tencent.com/document/product/248/45069)
物理专线|QCE/DC|[指标详情](https://cloud.tencent.com/document/product/248/45102)
专用通道|QCE/DCX|[指标详情](https://cloud.tencent.com/document/product/248/45101)
云硬盘|QCE/CBS|[指标详情](https://cloud.tencent.com/document/product/248/45411)
数据库SQL Server|QCE/SQLSERVER|[指标详情](https://cloud.tencent.com/document/product/248/45146)
数据库MariaDB|QCE/MARIADB|[指标详情](https://cloud.tencent.com/document/product/248/54397)
Elasticsearch|QCE/CES|[指标详情](https://cloud.tencent.com/document/product/248/45129)
CMQ 队列服务|QCE/CMQ(即将下线，不推荐)|[指标详情](https://cloud.tencent.com/document/product/248/45114)
CMQ 主题订阅|QCE/CMQTOPIC(即将下线，不推荐)|[指标详情](https://cloud.tencent.com/document/product/248/45113)
数据库PostgreSQL|QCE/POSTGRES|[指标详情](https://cloud.tencent.com/document/product/248/45105)
CKafka 实例|QCE/CKAFKA|[指标详情](https://cloud.tencent.com/document/product/248/45121)
Memcached |QCE/MEMCACHED|指标详情说明文档(待上线)
轻量应用服务器Lighthouse |QCE/LIGHTHOUSE|[指标详情](https://cloud.tencent.com/document/product/248/60127)
分布式数据库 TDSQL MySQL|QCE/TDMYSQL|[指标详情](https://cloud.tencent.com/document/product/248/54401)
弹性公网 IP|QCE/LB|[指标详情](https://cloud.tencent.com/document/product/248/45099)
消息队列RocketMQ版|QCE/TDMQ|[指标详情](https://cloud.tencent.com/document/product/248/51450#tdmq-rocketmq-.E7.89.88)
VPN 网关|QCE/VPNGW|[指标详情](https://cloud.tencent.com/document/product/248/45070)
VPN 通道|QCE/VPNX|[指标详情](https://cloud.tencent.com/document/product/248/45071)
CYNOSDB_MYSQL|QCE/CYNOSDB_MYSQL|[指标详情](https://cloud.tencent.com/document/product/248/45106)
云联网|QCE/VBC|[指标详情](https://cloud.tencent.com/document/product/248/75629)
数据传输 |QCE/DTS|[指标详情](https://cloud.tencent.com/document/product/248/82251)
专线网关 |QCE/DCG|[指标详情](https://cloud.tencent.com/document/product/248/45072)
全球应用加速|QCE/QAAP|[指标详情](https://cloud.tencent.com/document/product/248/45062)
Web应用防火墙 |QCE/WAF|[指标详情](https://cloud.tencent.com/document/product/248/48124)
负载均衡CLB(内网)|QCE/LB_PRIVATE|[指标详情](https://cloud.tencent.com/document/product/248/51899)

`后续会有更多的产品支持`

## 二、快速开始
### 1.构建
```shell
git clone https://github.com/tencentyun/tencentcloud-exporter.git
go build cmd/qcloud-exporter/qcloud_exporter.go
```
或从release列表获取预编译的二进制, 目前只提供linux-amd64
### 2. 定义产品实例配置
- 配置云API的`credential`认证信息
- 配置产品`products`指标、实例导出信息

如导出MongoDB所有指标所有实例

```yaml
credential:
  access_key: "access_key"            // 云API的SecretId
  secret_key: "secret_key"            // 云API的SecretKey
  region: "ap-guangzhou"              // 实例所在区域信息

products:
  - namespace: QCE/CMONGO             // 产品命名空间
    all_metrics: true                 // 导出支持的所有指标
    all_instances: true               // 导出region下的所有实例
    extra_labels: [InstanceName,Zone] // 将实例的字段作为指标的lables导出
```

### 3. 启动 Exporter

```bash
> qcloud_exporter --config.file "qcloud.yml"
```

访问 [http://127.0.0.1:9123/metrics](http://127.0.0.1:9123/metrics) 查看所有导出的指标




## 三、qcloud.yml配置详情
在git的`configs`里有支持产品的配置模版样例可参考
```yaml
credential:
  access_key: <YOUR_ACCESS_KEY>                  // 必须, 云API的SecretId
  secret_key: <YOUR_ACCESS_SECRET>               // 必须, 云API的SecretKey
  region: <REGION>                               // 必须, 实例所在区域信息

rate_limit: 15                                   // 腾讯云监控拉取指标数据限制, 官方默认限制最大20qps


// 整个产品纬度配置, 每个产品一个item
products:
  - namespace: QCE/CMONGO                        // 必须, 产品命名空间
    all_metrics: true                            // 常用, 推荐开启, 导出支持的所有指标
    all_instances: true                          // 常用, 推荐开启, 导出该region下的所有实例
    extra_labels: [InstanceName,Zone]            // 可选, 将实例的字段作为指标的lables导出
    only_include_metrics: [Inserts]              // 可选, 只导出这些指标, 配置时all_metrics失效
    exclude_metrics: [Reads]                     // 可选, 不导出这些指标
    only_include_instances: [cmgo-xxxxxxxx]      // 可选, 只导出这些实例id, 配置时all_instances失效
    exclude_instances: [cmgo-xxxxxxxx]           // 可选, 不导出这些实例id
    custom_query_dimensions:                     // 可选, 不常用, 自定义指标查询条件, 配置时all_instances,only_include_instances,exclude_instances失效, 用于不支持按实例纬度查询的指标
      - target: cmgo-xxxxxxxx
    statistics_types: [avg]                      // 可选, 拉取N个数据点, 再进行max、min、avg、last计算, 默认last取最新值
    period_seconds: 60                           // 可选, 指标统计周期
    range_seconds: 300                           // 可选, 选取时间范围, 开始时间=now-range_seconds, 结束时间=now
    delay_seconds: 60                            // 可选, 时间偏移量, 结束时间=now-delay_seconds
    metric_name_type: 1                          // 可选，导出指标的名字格式化类型, 1=大写转小写加下划线, 2=转小写; 默认2
    reload_interval_minutes: 60                   // 可选, 在all_instances=true时, 周期reload实例列表, 建议频率不要太频繁


// 单个指标纬度配置, 每个指标一个item
metrics:
  - tc_namespace: QCE/CMONGO                     // 产品命名空间, 同namespace
    tc_metric_name: Inserts                      // 云监控定义的指标名
    tc_metric_rename: Inserts                    // 导出指标的显示名
    tc_metric_name_type: 1                       // 可选，导出指标的名字格式化类型, 1=大写转小写加下划线, 2=转小写; 默认1
    tc_labels: [InstanceName]                    // 可选, 将实例的字段作为指标的lables导出
    tc_myself_dimensions:                        // 可选, 同custom_query_dimensions
    tc_statistics: [Avg]                         // 可选, 同statistics_types
    period_seconds: 60                           // 可选, 同period_seconds
    range_seconds: 300                           // 可选, 同range_seconds
    delay_seconds: 60                            // 可选, 同delay_seconds
```
特殊说明:
1. **custom_query_dimensions**  
   每个实例的纬度字段信息, 可从对应的云监控产品指标文档查询, 如mongo支持的纬度字段信息可由[云监控指标详情](https://cloud.tencent.com/document/product/248/45104#%E5%90%84%E7%BB%B4%E5%BA%A6%E5%AF%B9%E5%BA%94%E5%8F%82%E6%95%B0%E6%80%BB%E8%A7%88) 查询
2. **extra_labels**  
   每个导出metric的labels还额外上报实例对应的字段信息, 实例可选的字段列表可从对应产品文档查询, 如mongo实例支持的字段可从[实例查询api文档](https://cloud.tencent.com/document/product/240/38568) 获取, 目前只支持str、int类型的字段
3. **period_seconds**  
   每个指标支持的时间纬度统计, 一般支持60、300秒等, 具体可由对应产品的云监控产品指标文档查询, 如mongo可由[指标元数据查询](https://cloud.tencent.com/document/product/248/30351) , 假如不配置, 使用默认值(60), 假如该指标不支持60, 则自动使用该指标支持的最小值
4. **credential**  
   SecretId、SecretKey、Region可由环境变量获取
```bash
export TENCENTCLOUD_SECRET_ID="YOUR_ACCESS_KEY"
export TENCENTCLOUD_SECRET_KEY="YOUR_ACCESS_SECRET"
export TENCENTCLOUD_REGION="REGION"
```

5. **region**  
   地域可选值参考[地域可选值](https://cloud.tencent.com/document/api/248/30346#.E5.9C.B0.E5.9F.9F.E5.88.97.E8.A1.A8)
## 四、qcloud_exporter支持的命令行参数说明

命令行参数|说明|默认值
-------|----|-----
--web.listen-address|http服务的端口|9123
--web.telemetry-path|http访问的路径|/metrics
--web.enable-exporter-metrics|是否开启服务自身的指标导出, promhttp_\*, process_\*, go_*|false
--web.max-requests|最大同时抓取/metrics并发数, 0=disable|0
--config.file|产品实例指标配置文件位置|qcloud.yml
--log.level|日志级别|info


## 五、qcloud.yml样例
在git的configs里有支持产品的配置模版样例














