# GIPP: Go IP Proxy Pool

通过RESTful API，给其他爬虫程序提供**稳定**、**高效**的IP代理。

通过go的高并发，定期爬取大量免费的代理资源，进行质量筛选，并存储到Storage中。

## 组织架构

- IPProxy: ip代理对象，包括addr, port, type, region, https, speed等属性
- Spider: 免费代理资源爬取器，接口类型，不同的免费资源有不同的实现，例如ProxyListPlusSpider/IP181Spider等
- Inspector: 检验代理质量，包括时延、网速等等
- DBSession: 存储IPProxy的介质，例如MySQL、Mongo、Redis等等。
- Service: 提供api，获取可用ip代理

## 免费代理资源列表

## Usage
