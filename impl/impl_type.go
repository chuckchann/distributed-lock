package impl

const (
	REDIS = iota
	ETCD
	ZOOKEEPER
)

//current using type
var Type int
