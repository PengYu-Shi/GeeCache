package etcd

import (
	"GeeServer/tool"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"strings"
	"time"
)

var client *clientv3.Client

var currentNodes map[string]string

func Init(address []string) (err error) {
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		logrus.Error("Connect to etcd error: ", err)
	}
	return err
}
func GetConf(prefix string) {
	currentNodes = make(map[string]string)
	kv0 := clientv3.NewKV(client)
	getResp, _ := kv0.Get(context.TODO(), prefix, clientv3.WithPrefix())
	for _, kv := range getResp.Kvs {
		if !strings.HasPrefix(string(kv.Key), prefix) {
			log.Printf("警告：发现非预期键 %s，已跳过", string(kv.Key))
			continue
		}
		nodeID := strings.TrimPrefix(string(kv.Key), prefix)

		fmt.Println("Fount Node: ", nodeID, " ip: ", string(kv.Value))
		currentNodes[nodeID] = string(kv.Value)
	}
}

func WatchConf(prefix string) {
	watcher := clientv3.NewWatcher(client)
	watchChan := watcher.Watch(context.TODO(), prefix, clientv3.WithPrefix())
	go func() {
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				nodeID := strings.TrimPrefix(string(event.Kv.Key), prefix)

				switch event.Type {
				case clientv3.EventTypePut: // 新增/更新节点
					if event.IsCreate() {
						fmt.Println("Fount New Node: ", nodeID, " ip: ", string(event.Kv.Value))
					}
					currentNodes[nodeID] = string(event.Kv.Value)
					tool.HashMap.Add([]string{nodeID})

				case clientv3.EventTypeDelete: // 节点删除
					fmt.Println("Node Lost!! node : ", nodeID)
					tool.HashMap.Delete(nodeID)

					delete(currentNodes, nodeID)
				}
			}
		}
	}()

}

func GetKeyIp() (collectNodeList []string, err error) {
	if len(currentNodes) == 0 {
		return collectNodeList, fmt.Errorf("there is no node! ")
	}
	for _, v := range currentNodes {
		collectNodeList = append(collectNodeList, v)
	}
	return collectNodeList, nil
}
