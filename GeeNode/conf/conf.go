package conf

type AllConfig struct {
	Etcd  EtcdConfig  `ini:"etcd"`
	Node  NodeConfig  `ini:"node"`
	Mysql MysqlConfig `ini:"mysql"`
}

type EtcdConfig struct {
	Address string `ini:"address"`
	Prefix  string `ini:"nodePrefix"`
}

type NodeConfig struct {
	NodeIp        string `ini:"nodeIP"`
	FirstDataPath string `ini:"firstDataPath"`
	TailDataPath  string `ini:"tailDataPath"`
	LoadDuration  int    `ini:"dataDuration"`
	NodeFile      string `ini:"nodeFile"`
}

type MysqlConfig struct {
	DB      string `ini:"DB"`
	DbHost  string `ini:"DbHost"`
	DbPort  string `ini:"DbPort"`
	DbUser  string `ini:"DbUser"`
	DbPass  string `ini:"DbPass"`
	DbName  string `ini:"DbName"`
	Charset string `ini:"charset"`
}
