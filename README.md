# lru--improve

根据需求在golang官方提供的 [groupcache](https://github.com/golang/groupcache) 上做了一些改动

## 目的

在本地缓存1000000条下载过的图片链接，key为md5sum加密后的长度为16位的字符切片，value为空结构体。达到缓存资源利用率最优化。

## 改动

1. 将本地缓存key的类型为`[16]byte`；
2. 为缓存加锁，为缓存增加save字段，类型为`[]byte`；
3. 增加本地缓存持久化方法（定时加载缓存到本地文件以及读取本地文件到缓存）。

## 使用

1. 添加md5sum缓存

    ```go
    // 为本地增加链接缓存
    url := "http://www.google.com"
    newCache := lru.New(1000000)
    t1 := time.Now()
    
    for i := 0; i < 1000000; i++ {
        // 加密后的数据为长度为16位的字符切片
        newCache.Add(md5.Sum([]byte(url+strconv.Itoa(i))), struct{}{})
    }
    // 计算增加1000000条加密链接用时
    fmt.Println(time.Since(t1))
    ```

2. 获取md5sum缓存

    ```go

    if _, ok := newCache.Get(md5.Sum([]byte(url))); ok {
        fmt.Println("ok")
    }

    ```

3. 定时保存缓存到本地文件

    ```go

    go func() {
        ticker := time.NewTicker(time.Hour * time.Duration(cfg.ImageCache.Duration))
        for {
            select {
            case <-ticker.C:
                if err = newCache.SaveFile(logger, "fileName"); err != nil {
                    _ = level.Error(logger).Log("write file error", err)
                }
            }
        }

    }()
    ```

4. 加载本地文件到缓存

    ```go
    if err = imageUrlCache.LoadFile(logger, cfg.ImageCache.FileName); err != nil {
        _ = level.Error(logger).Log("load file to cache error", err)
    }
    
    ```
