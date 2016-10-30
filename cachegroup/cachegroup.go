package cachegroup

type CacheGroup []string

type CacheGroupManager struct {
	CacheGroups     []CacheGroup
	EndpointIndexes map[string][]int
}

func NewCacheGroupManager() *CacheGroupManager {
	cacheGroups := make([]CacheGroup, 0, 100)

	cgm := &CacheGroupManager{CacheGroups: cacheGroups, EndpointIndexes: make(map[string][]int)}
	return cgm
}

func CreateCacheGroup(endpoints ...string) CacheGroup {
	cg := make(CacheGroup, len(endpoints))
	for i := 0; i < len(endpoints); i++ {
		cg[i] = endpoints[i]
	}
	return cg
}

func (cgm *CacheGroupManager) AddCacheGroup(cgs ...CacheGroup) {
	for _, cg := range cgs {
		cgm.CacheGroups = append(cgm.CacheGroups, cg)
		for _, endpoints := range cg {
			indexes := cgm.EndpointIndexes[endpoints]
			indexes = append(indexes, len(cgm.CacheGroups)-1)
			cgm.EndpointIndexes[endpoints] = indexes
		}
	}
}

func (cgm *CacheGroupManager) GetGroupCacheIndexes(endpoint string) []int {
	return cgm.EndpointIndexes[endpoint]
}
