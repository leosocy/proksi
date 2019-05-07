# GIPP: Go IP Proxy Pool

[![Build Status](https://travis-ci.org/Leosocy/gipp.svg?branch=master)](https://travis-ci.org/Leosocy/gipp)
[![codecov](https://codecov.io/gh/Leosocy/gipp/branch/master/graph/badge.svg)](https://codecov.io/gh/Leosocy/gipp)

     _________        _______________   ________                               ________            ______
     __  ____/_____   ____  _/__  __ \  ___  __ \________________  ______  __  ___  __ \______________  /
     _  / __ _  __ \   __  / __  /_/ /  __  /_/ /_  ___/  __ \_  |/_/_  / / /  __  /_/ /  __ \  __ \_  / 
     / /_/ / / /_/ /  __/ /  _  ____/   _  ____/_  /   / /_/ /_>  < _  /_/ /   _  ____// /_/ / /_/ /  /  
     \____/  \____/   /___/  /_/        /_/     /_/    \____//_/|_| _\__, /    /_/     \____/\____//_/   
                                                                    /____/                               

> 支持以下模式

- 作为gate代理来自client请求，并根据proxy pool转发请求到最佳的代理。
- 作为datasource返回client请求的代理。

通过RESTful API，给其他爬虫程序提供**稳定**、**实时**、**高可用**的HTTP/HTTPS代理。

通过go的高并发，周期性爬取大量免费的代理资源，进行质量筛选，并存储到Storage中。

由于目前大部分网站都会重定向到https，所以如果代理不支持访问HTTPS，对需要使用代理的爬虫程序来说用处就不是很大了。

基于这种情况，GIPP **只会提供HTTP/HTTPS均支持** 的代理。并且会定期的对这些代理进行质量检查并打分，从而甄别出质量很高的代理。

## 组织架构

- Proxy: http(s)代理对象，包括ip, port, geo info, anonymity, latency, speed等属性。
- Spider: 免费代理资源爬取器, TODO: 从配置文件加载生成
- Checker: 检验代理质量，包括时延、网速等等，同时给代理打分。
- Storage: 存储Proxy的介质，例如MySQL、Mongo、Redis等等。
- Scheduler: 负责调度Spider,Checker,Storage之间的合作
- Service: 提供api，获取可用ip代理。

## 主要用到的开源包

- [colly](https://github.com/gocolly/colly)，用于发起请求，解析响应等等。

## 免费代理资源列表

## 启动

## API Usage

### proxies

|                                API                                | Method |             Description              |                       Args                        |  Try  |
| :---------------------------------------------------------------: | :----: | :----------------------------------: | :-----------------------------------------------: | :---: |
|           `http://localhost:8000/proxies?ipp=10&page=1`           |  GET   | 根据Score.Desc，返回指定页的10个代理 | `ipp`:一页返回n条记录，range(0, 50]  `page`:第n页 |       |
| `http://localhost:8000/proxies?ipp=10&page=1&geo.country_code=CN` |  GET   | 根据Geo信息的国家码返回`中国`的代理  |                  `geo.xxx`: xxx                   |

### uas
