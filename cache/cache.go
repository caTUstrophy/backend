package cache

type MapCache struct {
	Keys map[string]bool
}

type Cache interface {
	Get(key string) bool
	Set(key string)
}

func (dataObject MapCache) Get(key string) bool {
	_, ret := dataObject.Keys[key]
	return ret
}

func (dataObject MapCache) Set(key string) {
	dataObject.Keys[key] = true
}

func (dataObject MapCache) Unset(key string) {
	delete(dataObject.Keys, key)
}
