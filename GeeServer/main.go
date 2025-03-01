package main

import (
	"GeeServer/etcd"
	"GeeServer/geegin"
	"GeeServer/tool"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"time"
)

type AllConfig struct {
	Etcd    EtcdConfig    `ini:"etcd"`
	Content ContentConfig `ini:"content"`
}

type EtcdConfig struct {
	Address    string `ini:"address"`
	CollectKey string `ini:"collect_key"`
	Prefix     string `ini:"nodePrefix"`
}

type ContentConfig struct {
	Replicas int    `ini:"consistentHashReplicas"`
	GinPort  string `ini:"ginPort"`
}

func main() {
	var config = new(AllConfig)

	err := ini.MapTo(config, "conf/config.ini")
	if err != nil {
		logrus.Error("Get etcd config error: ", err)
		return
	}
	fmt.Println("address: ", config.Etcd.Address)
	time.Sleep(time.Second * 10)
	//连接etcd
	fmt.Println("----------------------------------------------Init Etcd----------------------------------------------")
	err = etcd.Init([]string{config.Etcd.Address})
	if err != nil {
		logrus.Error("init etcd error: ", err)
		return
	}

	//从etcd获取配置
	etcd.GetConf(config.Etcd.Prefix)
	nodeConf, err := etcd.GetKeyIp()
	if err != nil {
		fmt.Printf("get key ip err")
		return
	}
	fmt.Println("----------------------------------------------Start Watch Prefix :", config.Etcd.Prefix, "----------------------------------------------")

	etcd.WatchConf(config.Etcd.Prefix)
	//fmt.Println(string(nodeConf[0]))
	//time.Sleep(time.Hour)
	//配置一致性哈希
	fmt.Println("----------------------------------------------ConsistentHash----------------------------------------------")
	tool.New(config.Content.Replicas, nil)
	tool.HashMap.Add(nodeConf)

	//启动Gin
	fmt.Println("----------------------------------------------Init Gin----------------------------------------------")
	r := gin.Default()
	r = geegin.NewRouter(r)
	err = r.Run(config.Content.GinPort)
	if err != nil {
		return
	}

}
