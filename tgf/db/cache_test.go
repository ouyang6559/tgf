package db_test

import (
	"github.com/thkhxm/tgf/db"
	"reflect"
	"testing"
	"time"
)

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ 277949041
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/3/14
//***************************************************

func TestDefaultAutoCacheManager(t *testing.T) {
	cacheManager := db.NewDefaultAutoCacheManager[string, int64]("example")
	key := "1001"
	var setVal int64 = 10086
	val, err := cacheManager.Get(key)
	if err != nil {
		t.Errorf("[test] cache get error %v", err)
		return
	}
	//
	t.Logf("[test] cache first get key %v , val %v", key, val)

	cacheManager.Set(key, setVal)
	t.Logf("[test] cache set key %v , val %v ", key, setVal)
	val, err = cacheManager.Get(key)
	if err != nil {
		t.Errorf("[test] cache get error %v", err)
		return
	}
	t.Logf("[test] cache second get key %v , val %v", key, val)

	//first run
	//cache_test.go:26: [test] cache first get key 1001 , val 0
	//cache_test.go:29: [test] cache set key 1001 , val 0
	//cache_test.go:35: [test] cache second get key 1001 , val 10086

	//second run
	//cache_test.go:26: [test] cache first get key 1001 , val 10086
	//cache_test.go:29: [test] cache set key 1001 , val 10086
	//cache_test.go:35: [test] cache second get key 1001 , val 10086
}

func TestAddListItem(t *testing.T) {
	type args struct {
		key string
		val any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"1", args{key: "k1", val: 1}, false},
		{"2", args{key: "k1", val: 2}, false},
		{"3", args{key: "k1", val: 3}, false},
		{"4", args{key: "k1", val: 4}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.AddListItem(tt.args.key, tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("AddListItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetList(t *testing.T) {
	type args struct {
		key string
	}
	type testCase[Res any] struct {
		name string
		args args
		want []Res
	}
	tests := []testCase[string]{
		{"1", args{key: "k1"}, []string{"4", "3", "2", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := db.GetList[string](tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDel(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.Del(tt.args.key)
		})
	}
}

func TestDelNow(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.DelNow(tt.args.key)
		})
	}
}

func TestGetMap(t *testing.T) {
	//type args struct {
	//	key string
	//}
	//type testCase[Key cacheKey, Val any] struct {
	//	name        string
	//	args        args
	//	wantRes     map[Key]Val
	//	wantSuccess bool
	//}
	//tests := []testCase[ /* TODO: Insert concrete types here */ ]{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		gotRes, gotSuccess := db.GetMap(tt.args.key)
	//		if !reflect.DeepEqual(gotRes, tt.wantRes) {
	//			t.Errorf("GetMap() gotRes = %v, want %v", gotRes, tt.wantRes)
	//		}
	//		if gotSuccess != tt.wantSuccess {
	//			t.Errorf("GetMap() gotSuccess = %v, want %v", gotSuccess, tt.wantSuccess)
	//		}
	//	})
	//}
}

func TestPutMap(t *testing.T) {
	//type args[Key cacheKey, Val any] struct {
	//	key     string
	//	field   db.Key
	//	val     db.Val
	//	timeout time.Duration
	//}
	//type testCase[Key cacheKey, Val any] struct {
	//	name string
	//	args args[Key, Val]
	//}
	//tests := []testCase[ /* TODO: Insert concrete types here */ ]{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		db.PutMap(tt.args.key, tt.args.field, tt.args.val, tt.args.timeout)
	//	})
	//}
}

func TestSet(t *testing.T) {
	type args struct {
		key     string
		val     any
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.Set(tt.args.key, tt.args.val, tt.args.timeout)
		})
	}
}

func TestSetList(t *testing.T) {
	type args[Val any] struct {
		key     string
		l       []Val
		timeout time.Duration
	}
	type testCase[Val any] struct {
		name string
		args args[Val]
	}
	tests := []testCase[string]{
		{"1", args[string]{
			key:     "k1",
			l:       []string{"1", "1", "1", "1"},
			timeout: 0,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.SetList(tt.args.key, tt.args.l, tt.args.timeout)
		})
	}
}