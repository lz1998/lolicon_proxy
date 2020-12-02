# Lolicon Proxy

这是一个Pixiv插图漫画预加载的程序

背景：`https://i.pixiv.cat`插图漫画加载速度太慢

功能：自动从[lolicon](https://api.lolicon.app/)接口获取图片URL，并预先下载图片，在调用proxy时直接返回预先下载的图片，提升速度

## 获取APIKEY

根据[lolicon](https://api.lolicon.app/)说明，在tg申请apikey

## 下载lolicon_proxy

在[Actions](https://github.com/lz1998/lolicon_proxy/actions)中选择Linux或Windows，选择最新Result，并下载Artifacts

## 运行lolicon_proxy

### Windows

解压后直接双击

### Linux

先解压

```shell
chmod +x ./lolicon_proxy
PORT=18000 ./lolicon_proxy
```

**可以通过环境变量`PORT`设置端口，默认18000**

## 设置APIKEY和CACHE_COUNT

使用浏览器访问`http://localhost:18000/config?apikey=xxx&cache_count=10`，建议缓存数量不超过50，数量太大可能会卡

**可以在启动时通过环境变量`LOLICON_APIKEY`和`CACHE_COUNT`设置，运行后就不需要访问config接口**

## 获取图片

使用浏览器访问`http://localhost:18000/lolicon?r18=0&keyword=`，或前端/机器人等应用直接填写链接

**速度优化仅针对无`keyword`的情况，如果使用`keyword`参数，会直接调用API，速度较慢**

## 程序逻辑

数量检测：保存imageInfo到Sqlite(lolicon.db文件)，并标记used=0，使用后used标记为1，统计used=0的imageInfo数量

检测时间：程序启动时、config接口被调用后、每次请求图片前

自动下载：如果图片数量小于`CACHE_COUNT`，自动调用[lolicon](https://api.lolicon.app/)获取URL，并下载图片，保存到Sqlite
