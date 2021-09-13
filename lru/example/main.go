package main

import (
	"crypto/md5"
	"fmt"
	"github.com/go-kit/kit/log"
	"lru/lru"
	"os"
	"strconv"
	"time"
)

/*
 * @Author: Gpp
 * @File:   main.go
 * @Date:   2021/9/13 3:36 下午
 */

func main() {
	var logger log.Logger
	{
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(
			logger,
			"ts", log.DefaultTimestampUTC,
		)
	}

	//
	url := "http://www.google.com"
	newCache := lru.New(1000000)
	t1 := time.Now()

	for i := 0; i < 1000000; i++ {
		fmt.Println(md5.Sum([]byte(url + strconv.Itoa(i))))
		newCache.Add(md5.Sum([]byte(url+strconv.Itoa(i))), struct{}{})
	}
	fmt.Println(time.Since(t1))

	//if err := newCache.SaveFile(logger, "cache.dat"); err != nil {
	//	fmt.Println(err)
	//}

	if err := newCache.LoadFile(logger, "cache.dat"); err != nil {
		fmt.Println(err)
	}

	fmt.Println(time.Since(t1))
}
