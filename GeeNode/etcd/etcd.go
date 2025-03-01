package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Init(etcdAddr, prefix, port string) {
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{etcdAddr}})
	if err != nil {
		fmt.Println("failed to connect etcd")

	}
	lease := clientv3.NewLease(client)
	grantResp, _ := lease.Grant(context.TODO(), 30)
	leaseID := grantResp.ID

	// 节点注册（将租约与节点信息绑定）
	kv := clientv3.NewKV(client)
	key := fmt.Sprintf("%s%s", prefix, port)
	kv.Put(context.TODO(), key, port, clientv3.WithLease(leaseID))
	fmt.Println("success to put node, key: ", key, " value:", port)

	// 定期续约（每10秒一次）
	keepAliveCh, _ := lease.KeepAlive(context.TODO(), leaseID)
	go func() {
		for range keepAliveCh {
			// 续约成功，保持租约活性
		}
	}()
}
