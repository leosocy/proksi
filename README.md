# GIPP: Go IP Proxy Pool

[![Build Status](https://travis-ci.org/Leosocy/gipp.svg?branch=master)](https://travis-ci.org/Leosocy/gipp)
[![codecov](https://codecov.io/gh/Leosocy/gipp/branch/master/graph/badge.svg)](https://codecov.io/gh/Leosocy/gipp)

通过RESTful API，给其他爬虫程序提供**持久**、**稳定**、**高效**的IP代理。

通过go的高并发，定期爬取大量免费的代理资源，进行质量筛选，并存储到Storage中。

## 组织架构

- Proxy: http(s)代理对象，包括ip, port, geo info, anonymity, latency, speed等属性。
- Spider: 免费代理资源爬取器，接口类型，不同的免费资源有不同的实现，例如ProxyListPlusSpider/IP181Spider等。
- Checker: 检验代理质量，包括时延、网速等等。
- Storage: 存储Proxy的介质，例如MySQL、Mongo、Redis等等。
- Service: 提供api，获取可用ip代理。

## 免费代理资源列表

## 启动

## API Usage

### proxies

|                                API                                | Method |             Description              |                       Args                        |  Try  |
| :---------------------------------------------------------------: | :----: | :----------------------------------: | :-----------------------------------------------: | :---: |
|           `http://localhost:8000/proxies?ipp=10&page=1`           |  GET   | 根据Score.Desc，返回指定页的10个代理 | `ipp`:一页返回n条记录，range(0, 50]  `page`:第n页 |       |
| `http://localhost:8000/proxies?ipp=10&page=1&geo.country_code=CN` |  GET   | 根据Geo信息的国家码返回`中国`的代理  |                  `geo.xxx`: xxx                   |

### uas
