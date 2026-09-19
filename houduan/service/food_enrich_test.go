package service

import (
	"testing"

	"wenlv-backend/model"
)

func TestMatchFoodPOI(t *testing.T) {
	candidates := []model.POIInfo{
		{Name: "肖家河美味家常面(肖家河北街店)", Type: "餐饮服务;中餐厅;中餐厅"},
		{Name: "担担面(总店)", Type: "餐饮服务;中餐厅;特色/地方风味餐厅"},
		{Name: "担担面(文殊坊店)", Type: "餐饮服务;中餐厅;中餐厅"},
	}
	got, ok := matchFoodPOI(candidates, "担担面")
	if !ok {
		t.Fatal("应当匹配到同名门店")
	}
	if got.Name != "担担面(总店)" {
		t.Errorf("应优先取名称包含菜名的门店, got %q", got.Name)
	}
	if _, ok := matchFoodPOI([]model.POIInfo{{Name: "肖家河美味家常面", Type: "餐饮服务;中餐厅;中餐厅"}}, "担担面"); ok {
		t.Error("无同名门店时必须返回未匹配,不能随便取一个餐厅")
	}
	if _, ok := matchFoodPOI([]model.POIInfo{{Name: "陈麻婆豆腐(旗舰店)", Type: "购物服务;超市"}}, "麻婆豆腐"); ok {
		t.Error("非餐饮服务类型应被排除")
	}
	if _, ok := matchFoodPOI(nil, "火锅"); ok {
		t.Error("空候选应返回未匹配")
	}
}

func TestFoodIsUsablePOI(t *testing.T) {
	cases := map[string]bool{
		"餐饮服务;中餐厅;四川菜(川菜)": true,
		"购物服务;超级市场;超市":     false,
		"餐饮服务;快餐厅;快餐厅":     true,
		"":                 false,
	}
	for in, want := range cases {
		if got := foodIsUsablePOI(in); got != want {
			t.Errorf("foodIsUsablePOI(%q)=%v, want %v", in, got, want)
		}
	}
}

func TestIsEmptyFoodField(t *testing.T) {
	f := model.Food{}
	if !isEmptyFoodField(f, "flavor") {
		t.Error("空记录的风味字段应视为空")
	}
	f2 := model.Food{Flavor: "麻辣鲜香", Images: "/images/placeholder-food.jpg"}
	if isEmptyFoodField(f2, "flavor") {
		t.Error("已有值不应视为空")
	}
	if !isEmptyFoodField(f2, "images") {
		t.Error("占位封面应视为空(可被真实封面替换)")
	}
	if !isEmptyFoodField(f2, "unknown_field") {
		t.Error("未知字段默认视为空")
	}
}
