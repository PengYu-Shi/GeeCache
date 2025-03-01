package main

import (
	"GeeCacheNode/DB"
	"GeeCacheNode/conf"
	"GeeCacheNode/etcd"
	"GeeCacheNode/gee"
	"net/http"

	//"GeeCacheNode/gee/snapshot"

	//"GeeCacheNode/gee/snapshot"
	"GeeCacheNode/model"
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"gorm.io/gorm"
	"log"
	//"strconv"
	"time"
)

func createGroup(duration int) *gee.Group {
	db := DB.GetDB()

	return gee.NewGroup("scores", 2<<19, time.Duration(duration)*time.Minute, gee.GetterFunc(
		func(key string) ([]byte, error) {
			var score model.Score
			result := db.Where("`key` = ?", key).First(&score)
			if result.Error != nil {
				if errors.Is(result.Error, gorm.ErrRecordNotFound) {
					return nil, nil // 或者返回一个自定义的错误
				}
				return nil, result.Error
			}
			log.Println("[MySQL]  hit")
			return []byte(score.Score), nil
		}))
}

//func createGroup(getter gee.GetterFunc) *gee.Group {
//	return gee.NewGroup("scores", 2<<10, gee.GetterFunc(
//		getter))
//}

func main() {
	var config = new(conf.AllConfig)
	err := ini.MapTo(config, "./conf/conf.ini")

	if err != nil {
		return
	}
	_, err = DB.InitDB(config.Mysql)
	if err != nil {
		fmt.Println("Init Database error: ", err)
	}

	//err = DB.InsertData(db)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	////}
	//fmt.Println(config.Node.FirstDataPath + config.Node.NodeIp + config.Node.TailDataPath)
	//time.Sleep(time.Hour)
	//snapshot.SetRoot(config.Node.FirstDataPath + config.Node.NodeFile + config.Node.TailDataPath)
	// 注册 pprof HTTP 接口
	go func() {
		log.Println("Starting pprof server on :6666")
		log.Println(http.ListenAndServe(":6666", nil)) // 启动 pprof 服务，默认路径为 /debug/pprof
	}()
	createGroup(config.Node.LoadDuration)
	fmt.Println("address: ", config.Etcd.Address)
	go gee.Init("127.0.0.1:9005")
	etcd.Init(config.Etcd.Address, config.Etcd.Prefix, "127.0.0.1:9005")
	for {
		time.Sleep(time.Hour)
	}
}
