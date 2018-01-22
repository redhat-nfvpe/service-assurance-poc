package cacheutil

type cache interface {
	GetItems() map[string]*interface{}
	Put(key string)
	Size() int
	GetShard() *interface{}
}
type shardedcache interface {
	GetDataItems(key string) *interface{}
	SetDataItems(data interface{})
}
