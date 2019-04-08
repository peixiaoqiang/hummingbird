# Code Structure
## etc
* 10-macvlan.conf：kubernetes macvlan cni 示例配置
* ipallocator.conf：ipallocator server 示例配置
* taskmanager.conf：spark task manager 示例配置

## pkg
* network/allocator：ipallocator server 代码
* network/cni：kubernetes cni 客户端代码
* spark：kubernetes spark task 代码
* storage：公共存储代码
* taskmanager：task manager 代码

## spark
kubernetes spark history 启动代码

## spark-on-kubernetes
spark on kubernetes 方案相关脚本和 dockerfile

## taskmanager
task manager 启动代码

## tools
一些工具类，包括各个模块的 dockerfile、测试脚本和 notebook 整合方案代码

