package cachegroup_test

import (
	"reflect"
	"testing"

	"github.com/221bytes/negroniredis/cachegroup"
)

func Test(t *testing.T) {
	cg0 := cachegroup.CreateCacheGroup("test", "tolo", "yolo")
	cg1 := cachegroup.CreateCacheGroup("tata", "tolo", "yolo")
	cg2 := cachegroup.CreateCacheGroup("toto", "tolo", "yolo")
	cgm := cachegroup.NewCacheGroupManager()
	cgm.AddCacheGroup(cg0, cg1, cg2)
	test := []int{0}
	v := cgm.GetGroupCacheIndexes("test")
	if !reflect.DeepEqual(v, test) {
		t.Error(
			"expected", test,
			"got", v,
		)
	}
}
