# IntelliProxy: Provide durable, real-time, high-quality proxies as a middleman or datasource server

[![Build Status](https://travis-ci.org/Leosocy/IntelliProxy.svg?branch=master)](https://travis-ci.org/Leosocy/IntelliProxy)
[![codecov](https://codecov.io/gh/Leosocy/IntelliProxy/branch/master/graph/badge.svg)](https://codecov.io/gh/Leosocy/IntelliProxy)

The client can simply use `IntelliProxy` as a proxy server to achieve random ip access to the target host. `IntelliProxy` acts as a middleman to forward client requests to real proxy servers according to certain strategies.
![middleman](https://blog-images-1257621236.cos.ap-shanghai.myqcloud.com/IntelliProxy-MiddlemanServer-High-Compress.gif)
Or the client can use `IntelliProxy` as a data source to request the required proxy through the RESTful API.

## 组织架构

- Proxy: http(s)代理对象，包括ip, port, geo info, anonymity, latency, speed等属性。
- Spider: 免费代理资源爬取器。
- Checker: 检验代理质量，包括时延、网速等等，同时给代理打分。
- Storage: 存储Proxy的介质，例如InMemory、MySQL、Mongo、Redis等等。
- Scheduler: 负责调度Spider, Checker, Storage之间的合作。
- Service  
  - middleman: client可以直接将代理服务器指向middleman监听的端口，IntelliProxy会选出最佳的代理服务器转发出去。
  - datasource: 提供RESTful API，支持查询符合条件的proxy。

## 主要用到的开源包

- [colly](https://github.com/gocolly/colly)，用于发起请求，解析响应等等。

## 免费代理资源列表

## 启动

## Usage

### middleman

### datasource

|                                API                                | Method |             Description              |                       Args                        |  Try  |
| :---------------------------------------------------------------: | :----: | :----------------------------------: | :-----------------------------------------------: | :---: |
|           `http://localhost:8000/proxies?ipp=10&page=1`           |  GET   | 根据Score.Desc，返回指定页的10个代理 | `ipp`:一页返回n条记录，range(0, 50]  `page`:第n页 |       |
| `http://localhost:8000/proxies?ipp=10&page=1&geo.country_code=CN` |  GET   | 根据Geo信息的国家码返回`中国`的代理  |                  `geo.xxx`: xxx                   |

## TODO List

- [ ] Spiders支持从config文件加载
