package service

import (
	"testing"

	"wenlv-backend/model"
)

func TestDishesForCategory(t *testing.T) {
	foods := []model.Food{
		{ID: 1, NameZH: "麻婆豆腐", Tags: "川菜,经典名菜"},
		{ID: 2, NameZH: "担担面", Tags: "小吃,面食"},
		{ID: 3, NameZH: "兔头", Tags: "夜宵,凉菜"},
		{ID: 4, NameZH: "甜水面", Tags: "面食"},
	}
	got := dishesForCategory(foods, []string{"川菜"})
	if len(got) != 1 || got[0].NameZH != "麻婆豆腐" {
		t.Fatalf("dishesForCategory(川菜) 应只返回麻婆豆腐, got %+v", got)
	}

	// 包含匹配:标签「一个人的火锅」应能被关键词「火锅」命中
	hotpot := []model.Food{{ID: 9, NameZH: "火锅", Tags: "一个人的火锅"}}
	if got := dishesForCategory(hotpot, []string{"火锅"}); len(got) != 1 {
		t.Fatalf("包含匹配应命中, got %+v", got)
	}

	// 多关键词为「或」关系
	if got := dishesForCategory(foods, []string{"夜宵", "川菜"}); len(got) != 2 {
		t.Fatalf("多关键词应命中 2 条(麻婆豆腐/兔头), got %+v", got)
	}

	// 空关键词不命中任何菜品
	if got := dishesForCategory(foods, nil); len(got) != 0 {
		t.Fatalf("无关键词应返回空, got %+v", got)
	}
	if got := dishesForCategory(foods, []string{"", "  "}); len(got) != 0 {
		t.Fatalf("空串关键词应返回空, got %+v", got)
	}

	// collectCategoryDishNames:命中菜品的 name_zh 名单,去重、受 max 限制,用于注入 LLM 提示词
	names := collectCategoryDishNames(foods, []string{"夜宵", "川菜"}, 8)
	if len(names) != 2 || names[0] != "麻婆豆腐" || names[1] != "兔头" {
		t.Fatalf("菜品名单错误, got %+v", names)
	}
	if capped := collectCategoryDishNames(foods, []string{"夜宵", "川菜"}, 1); len(capped) != 1 || capped[0] != "麻婆豆腐" {
		t.Fatalf("名单 max 限制错误, got %+v", capped)
	}
	if got := collectCategoryDishNames(foods, []string{"火锅"}, 8); len(got) != 0 {
		t.Fatalf("无命中应返回空名单, got %+v", got)
	}
}

func TestCollectCategoryImages(t *testing.T) {
	// 轮询(round-robin)语义:按菜品顺序,每轮从每道菜各取 1 张,
	// 直到取满 limit 或所有菜品的图都被取完(避免单个菜品垄断图集)。
	// 同一道菜内先取 gallery_images 再取 images;URL 全局去重、跳过占位封面。
	foods := []model.Food{
		{ID: 1, NameZH: "麻婆豆腐", Tags: "川菜", GalleryImages: "a1.jpg,a2.jpg"},
		{ID: 2, NameZH: "宫保鸡丁", Tags: "川菜", GalleryImages: "b1.jpg", Images: "b2.jpg"},
		{ID: 3, NameZH: "回锅肉", Tags: "川菜", Images: "c1.jpg,c2.jpg"},
		{ID: 4, NameZH: "担担面", Tags: "小吃", GalleryImages: "x1.jpg,x2.jpg"},
	}

	// 3 道菜各 2 张、limit=6:应为跨菜品交替,而非菜 A 的 6 张(菜 A 最多 2 张)
	got := collectCategoryImages(foods, []string{"川菜"}, 6)
	want := []string{"a1.jpg", "b1.jpg", "c1.jpg", "a2.jpg", "b2.jpg", "c2.jpg"}
	if len(got) != len(want) {
		t.Fatalf("collectCategoryImages 长度=%d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("collectCategoryImages 轮询顺序错误, got %v, want %v", got, want)
		}
	}

	// limit 截断:取到第 1 轮结束(每道菜各 1 张)即停,而非把菜 A 取满
	truncated := collectCategoryImages(foods, []string{"川菜"}, 3)
	wantTrunc := []string{"a1.jpg", "b1.jpg", "c1.jpg"}
	if len(truncated) != len(wantTrunc) {
		t.Fatalf("limit 截断长度错误, got %v", truncated)
	}
	for i := range wantTrunc {
		if truncated[i] != wantTrunc[i] {
			t.Fatalf("limit 截断内容错误, got %v, want %v", truncated, wantTrunc)
		}
	}

	// 去重:菜 2 的 gallery 与菜 1 的 images 同为 b.jpg,只保留首次出现;
	// 去重跳过不占用该轮的"每菜 1 张"名额
	dup := []model.Food{
		{ID: 1, NameZH: "麻婆豆腐", Tags: "川菜", GalleryImages: "a.jpg", Images: "b.jpg"},
		{ID: 2, NameZH: "宫保鸡丁", Tags: "川菜", GalleryImages: "b.jpg,d.jpg"},
	}
	dedup := collectCategoryImages(dup, []string{"川菜"}, 6)
	wantDedup := []string{"a.jpg", "b.jpg", "d.jpg"}
	if len(dedup) != len(wantDedup) {
		t.Fatalf("去重后长度错误, got %v", dedup)
	}
	for i := range wantDedup {
		if dedup[i] != wantDedup[i] {
			t.Fatalf("去重结果错误, got %v, want %v", dedup, wantDedup)
		}
	}

	if got := collectCategoryImages(foods, []string{"川菜"}, 0); len(got) != 0 {
		t.Fatalf("limit<=0 应返回空, got %v", got)
	}
}

func TestBuildCategorySections(t *testing.T) {
	secs := []model.CategorySection{
		{Title: "起源与特点", Text: "第一段正文。"},
		{Title: "味型与技法", Text: "第二段正文。"},
		{Title: "   ", Text: "缺标题应被丢弃。"},
		{Title: "代表菜品", Text: "   "},
		{Title: "去哪里吃", Text: "第三段正文。"},
	}

	got := buildCategorySections(secs, []string{"img1.jpg", "img2.jpg"})
	if len(got) != 3 {
		t.Fatalf("缺 title/text 的段落应被丢弃, got %+v", got)
	}
	wantImgs := []string{"img1.jpg", "img2.jpg", "img2.jpg"}
	for i, sec := range got {
		if sec.Image != wantImgs[i] {
			t.Fatalf("第 %d 段配图=%q, want %q(图片不足应复用最后一张)", i+1, sec.Image, wantImgs[i])
		}
		if sec.Title == "" || sec.Text == "" {
			t.Fatalf("第 %d 段 title/text 不应为空: %+v", i+1, sec)
		}
	}

	// 无图时不应 panic,配图留空
	noImg := buildCategorySections(secs, nil)
	if len(noImg) != 3 {
		t.Fatalf("无图时仍应保留 3 段, got %d", len(noImg))
	}
	for _, sec := range noImg {
		if sec.Image != "" {
			t.Fatalf("无图时配图应为空, got %q", sec.Image)
		}
	}

	if got := buildCategorySections(nil, []string{"img1.jpg"}); len(got) != 0 {
		t.Fatalf("空输入应返回空, got %+v", got)
	}
}
