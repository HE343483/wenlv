package service

import (
	"testing"

	"wenlv-backend/model"
)

func TestDistanceMeters(t *testing.T) {
	// 春熙路 → 天府广场,实际约 1.1km
	got := distanceMeters(104.0810, 30.6570, 104.0657, 30.6573)
	if got < 1200 || got > 1800 {
		t.Errorf("distanceMeters=%v, 期望 1200~1800 米", got)
	}
	if distanceMeters(0, 0, 0, 0) != 0 {
		t.Error("同点距离应为 0")
	}
}

func TestBuildDetailSections(t *testing.T) {
	images := []string{"cover.jpg", "g1.jpg", "g2.jpg"}
	llm := []model.ScenicDetailSection{
		{Title: "街区沿革", Text: "甲"},
		{Title: "今日风貌", Text: "乙"},
		{Title: "第三段", Text: "丙"},
	}
	got := buildDetailSections(llm, images)
	if len(got) != 3 {
		t.Fatalf("段落数=%d, want 3", len(got))
	}
	if got[0].Image != "g1.jpg" || got[1].Image != "g2.jpg" {
		t.Errorf("配图应从第 2 张开始配对, got %q/%q", got[0].Image, got[1].Image)
	}
	if got[2].Image != "g1.jpg" {
		t.Errorf("图片不足的段落应复用第 2 张图兜底, got %q", got[2].Image)
	}
	if got[0].Title != "街区沿革" || got[0].Text != "甲" {
		t.Errorf("段落内容被改写: %+v", got[0])
	}
}

func TestBuildDetailSectionsFillsEverySection(t *testing.T) {
	images := []string{"cover.jpg", "g1.jpg", "g2.jpg"}
	llm := []model.ScenicDetailSection{
		{Title: "一", Text: "甲"},
		{Title: "二", Text: "乙"},
		{Title: "三", Text: "丙"},
		{Title: "四", Text: "丁"},
	}
	got := buildDetailSections(llm, images)
	if len(got) != 4 {
		t.Fatalf("段落数=%d, want 4", len(got))
	}
	for i, sec := range got {
		if sec.Image == "" {
			t.Errorf("第 %d 段配图为空, 期望兜底复用", i)
		}
	}
	if got[0].Image != "g1.jpg" || got[1].Image != "g2.jpg" {
		t.Errorf("前两段应各不相同且从第 2 张开始, got %q/%q", got[0].Image, got[1].Image)
	}
	if got[0].Image == got[1].Image {
		t.Errorf("前两段配图不应相同, got %q", got[0].Image)
	}
	if got[2].Image != "g1.jpg" || got[3].Image != "g1.jpg" {
		t.Errorf("后续段落应复用第 2 张图, got %q/%q", got[2].Image, got[3].Image)
	}
	if len(images) != 3 {
		t.Fatalf("入参切片被修改: %v", images)
	}
}

func TestBuildDetailSectionsEmptyImages(t *testing.T) {
	llm := []model.ScenicDetailSection{{Title: "一", Text: "甲"}, {Title: "二", Text: "乙"}}
	got := buildDetailSections(llm, nil)
	if len(got) != 2 {
		t.Fatalf("段落数=%d, want 2", len(got))
	}
	for i, sec := range got {
		if sec.Image != "" {
			t.Errorf("无图时第 %d 段应维持空图(前端占位), got %q", i, sec.Image)
		}
	}
}

func TestClampMaxSections(t *testing.T) {
	cases := []struct {
		images int
		want   int
	}{
		{images: 0, want: 2},
		{images: 1, want: 2},
		{images: 2, want: 2},
		{images: 3, want: 2},
		{images: 5, want: 4},
		{images: 7, want: 4},
	}
	for _, tc := range cases {
		if got := clampMaxSections(tc.images); got != tc.want {
			t.Errorf("clampMaxSections(%d)=%d, want %d", tc.images, got, tc.want)
		}
	}
}

func TestCommonsFileAcceptable(t *testing.T) {
	names := []string{"青城山", "Mount Qingcheng"}
	cases := []struct {
		name   string
		title  string
		rawURL string
		width  int
		want   bool
	}{
		{
			name:   "中文文件名包含景点名且达标",
			title:  "File:青城山 山门.jpg",
			rawURL: "https://upload.wikimedia.org/wikipedia/commons/a/ab/Qingchengshan.jpg",
			width:  1600,
			want:   true,
		},
		{
			name:   "英文文件名包含英文景点名且达标",
			title:  "File:Mount Qingcheng trail.jpg",
			rawURL: "https://upload.wikimedia.org/wikipedia/commons/a/ab/Mount_Qingcheng.jpg",
			width:  1200,
			want:   true,
		},
		{
			name:   "英文同名但非景点名(Qingchengshan 连写)应拒绝",
			title:  "File:Qingchengshan Railway Station.jpg",
			rawURL: "https://upload.wikimedia.org/wikipedia/commons/a/ab/Station.jpg",
			width:  1200,
			want:   false,
		},
		{
			name:   "文件名不含景点名应拒绝",
			title:  "File:Chengdu skyline.jpg",
			rawURL: "https://upload.wikimedia.org/wikipedia/commons/a/ab/Skyline.jpg",
			width:  1600,
			want:   false,
		},
		{
			name:   "宽度不足应拒绝",
			title:  "File:青城山 山门.jpg",
			rawURL: "https://upload.wikimedia.org/wikipedia/commons/a/ab/Q.jpg",
			width:  500,
			want:   false,
		},
		{
			name:   "非 jpg/jpeg/png 应拒绝",
			title:  "File:青城山 平面图.svg",
			rawURL: "https://upload.wikimedia.org/wikipedia/commons/a/ab/Q.svg",
			width:  1200,
			want:   false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := commonsFileAcceptable(tc.title, tc.rawURL, tc.width, names); got != tc.want {
				t.Errorf("commonsFileAcceptable(%q, %q, %d)=%v, want %v",
					tc.title, tc.rawURL, tc.width, got, tc.want)
			}
		})
	}
}

func TestAppendEstimatedField(t *testing.T) {
	if got := appendEstimatedField("", "ticket_price"); got != "ticket_price" {
		t.Errorf("空串应直接写入, got %q", got)
	}
	if got := appendEstimatedField("ticket_price", "ticket_price"); got != "ticket_price" {
		t.Errorf("重复字段不应重复追加, got %q", got)
	}
	if got := appendEstimatedField("ticket_price", "recommend_hours"); got != "ticket_price,recommend_hours" {
		t.Errorf("追加顺序错误, got %q", got)
	}
}

func TestSurvivingEstimated(t *testing.T) {
	cases := []struct {
		name      string
		existing  string
		llmFields []string
		updates   map[string]any
		want      string
	}{
		{
			name:      "被非 Force 保护删掉的字段不登记",
			existing:  "",
			llmFields: []string{"ticket_price", "recommend_hours"},
			updates:   map[string]any{"ticket_price": "免费"},
			want:      "ticket_price",
		},
		{
			name:      "保留历史登记并追加本次写入字段",
			existing:  "yearly_visitors",
			llmFields: []string{"ticket_price"},
			updates:   map[string]any{"ticket_price": "免费"},
			want:      "yearly_visitors,ticket_price",
		},
		{
			name:      "全部字段被删时返回原值",
			existing:  "yearly_visitors",
			llmFields: []string{"ticket_price", "recommend_hours"},
			updates:   map[string]any{"address": "成都市武侯区"},
			want:      "yearly_visitors",
		},
		{
			name:      "原有登记为空且字段全被删时返回空串",
			existing:  "",
			llmFields: []string{"ticket_price"},
			updates:   map[string]any{},
			want:      "",
		},
		{
			name:      "已登记字段不重复追加",
			existing:  "ticket_price",
			llmFields: []string{"ticket_price"},
			updates:   map[string]any{"ticket_price": "免费"},
			want:      "ticket_price",
		},
		{
			name:      "llmFields 为空切片时返回原值",
			existing:  "ticket_price",
			llmFields: []string{},
			updates:   map[string]any{"ticket_price": "免费"},
			want:      "ticket_price",
		},
		{
			name:      "llmFields 为 nil 时返回原值",
			existing:  "ticket_price",
			llmFields: nil,
			updates:   map[string]any{"ticket_price": "免费"},
			want:      "ticket_price",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := survivingEstimated(tc.existing, tc.llmFields, tc.updates); got != tc.want {
				t.Errorf("survivingEstimated(%q, %v, %v)=%q, want %q",
					tc.existing, tc.llmFields, tc.updates, got, tc.want)
			}
		})
	}
}

func TestFilterEstimated(t *testing.T) {
	base := []estimatedKV{
		{Field: "ticket_price", Value: "免费"},
		{Field: "open_hours", Value: "8:00-18:00"},
		{Field: "recommend_hours", Value: "2-3小时"},
	}
	cases := []struct {
		name  string
		est   []estimatedKV
		taken map[string]any
		want  []string // 期望保留下来的字段名,按原顺序
	}{
		{
			name:  "taken 命中 open_hours 时过滤该字段",
			est:   base,
			taken: map[string]any{"open_hours": "09:00-17:00"},
			want:  []string{"ticket_price", "recommend_hours"},
		},
		{
			name:  "taken 命中多个字段时全部过滤",
			est:   base,
			taken: map[string]any{"ticket_price": "免费", "recommend_hours": "2-3小时"},
			want:  []string{"open_hours"},
		},
		{
			name:  "taken 为空 map 时全部保留",
			est:   base,
			taken: map[string]any{},
			want:  []string{"ticket_price", "open_hours", "recommend_hours"},
		},
		{
			name:  "taken 为 nil 时全部保留",
			est:   base,
			taken: nil,
			want:  []string{"ticket_price", "open_hours", "recommend_hours"},
		},
		{
			name:  "taken 只有无关键时全部保留",
			est:   base,
			taken: map[string]any{"address": "成都市武侯区"},
			want:  []string{"ticket_price", "open_hours", "recommend_hours"},
		},
		{
			name:  "est 为空时返回空",
			est:   nil,
			taken: map[string]any{"open_hours": "09:00-17:00"},
			want:  nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := append([]estimatedKV(nil), tc.est...)
			got := filterEstimated(tc.est, tc.taken)
			if len(got) != len(tc.want) {
				t.Fatalf("保留字段数=%d, want %d, got %+v", len(got), len(tc.want), got)
			}
			for i, want := range tc.want {
				if got[i].Field != want {
					t.Errorf("第 %d 个保留字段=%q, want %q", i, got[i].Field, want)
				}
			}
			// 不得修改入参切片本身
			if len(tc.est) != len(before) {
				t.Fatalf("入参切片长度被修改: %d -> %d", len(before), len(tc.est))
			}
			for i := range before {
				if tc.est[i] != before[i] {
					t.Errorf("入参切片第 %d 项被修改: %+v -> %+v", i, before[i], tc.est[i])
				}
			}
		})
	}
}

func TestIsUsablePOI(t *testing.T) {
	cases := []struct {
		name    string
		poiType string
		want    bool
	}{
		{
			name:    "春熙路步行街(特色商业街)可用",
			poiType: "购物服务;特色商业街;特色商业街",
			want:    true,
		},
		{
			name:    "春熙路(热点地名)不可用",
			poiType: "地名地址信息;热点地名;热点地名",
			want:    false,
		},
		{
			name:    "春熙路(道路名)不可用",
			poiType: "地名地址信息;交通地名;道路名",
			want:    false,
		},
		{
			name:    "春熙路(地铁站)不可用",
			poiType: "交通设施服务;地铁站;地铁站",
			want:    false,
		},
		{
			name:    "风景名胜景点可用",
			poiType: "风景名胜;风景名胜;景点",
			want:    true,
		},
		{
			name:    "空字符串可用",
			poiType: "",
			want:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUsablePOI(tc.poiType); got != tc.want {
				t.Errorf("isUsablePOI(%q)=%v, want %v", tc.poiType, got, tc.want)
			}
		})
	}
}

func TestIsUsablePOIName(t *testing.T) {
	cases := []struct {
		name string
		poi  string
		want bool
	}{
		{
			name: "凤凰湖湿地公园-公共厕所不可用",
			poi:  "凤凰湖湿地公园-公共厕所",
			want: false,
		},
		{
			name: "凤凰湖旅游景区可用",
			poi:  "凤凰湖旅游景区",
			want: true,
		},
		{
			name: "凤凰湖湿地公园地面停车场(出入口)不可用",
			poi:  "凤凰湖湿地公园地面停车场(出入口)",
			want: false,
		},
		{
			name: "成都大熊猫繁育研究基地可用",
			poi:  "成都大熊猫繁育研究基地",
			want: true,
		},
		{
			name: "春熙路步行街可用",
			poi:  "春熙路步行街",
			want: true,
		},
		{
			name: "空字符串可用",
			poi:  "",
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUsablePOIName(tc.poi); got != tc.want {
				t.Errorf("isUsablePOIName(%q)=%v, want %v", tc.poi, got, tc.want)
			}
		})
	}
}

func TestHasBusinessUpdate(t *testing.T) {
	cases := []struct {
		name    string
		updates map[string]any
		want    bool
	}{
		{
			name:    "空 map 不算业务变更",
			updates: map[string]any{},
			want:    false,
		},
		{
			name:    "nil map 不算业务变更",
			updates: nil,
			want:    false,
		},
		{
			name: "只含三个溯源键不算业务变更",
			updates: map[string]any{
				"estimated_fields": "ticket_price",
				"data_source":      "amap",
				"data_updated_at":  "2026-09-18T00:00:00Z",
			},
			want: false,
		},
		{
			name:    "含 address 算业务变更",
			updates: map[string]any{"address": "成都市武侯区"},
			want:    true,
		},
		{
			name:    "含 gallery_images 算业务变更",
			updates: map[string]any{"gallery_images": "https://oss/a.jpg,https://oss/b.jpg"},
			want:    true,
		},
		{
			name:    "含 amap_poi_id 算业务变更",
			updates: map[string]any{"amap_poi_id": "B001"},
			want:    true,
		},
		{
			name: "业务键与溯源键混合算业务变更",
			updates: map[string]any{
				"open_hours":       "08:00-18:00",
				"data_source":      "amap",
				"estimated_fields": "",
			},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasBusinessUpdate(tc.updates); got != tc.want {
				t.Errorf("hasBusinessUpdate(%+v)=%v, want %v", tc.updates, got, tc.want)
			}
		})
	}
}

func TestIsPlaceholderCover(t *testing.T) {
	cases := []struct {
		name  string
		cover string
		want  bool
	}{
		{name: "空串视为没有封面", cover: "", want: true},
		{name: "纯空白视为没有封面", cover: "   ", want: true},
		{name: "永陵占位路径视为没有封面", cover: "/images/placeholder-yongling.jpg", want: true},
		{name: "数字占位路径视为没有封面", cover: "/images/placeholder-1.jpg", want: true},
		{
			name:  "真实 OSS 地址不是占位",
			cover: "https://wenlv-tdx.oss-cn-chengdu.aliyuncs.com/scenic/21/g1.jpg",
			want:  false,
		},
		{name: "相对上传路径不是占位", cover: "/uploads/a.jpg", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isPlaceholderCover(tc.cover); got != tc.want {
				t.Errorf("isPlaceholderCover(%q)=%v, want %v", tc.cover, got, tc.want)
			}
		})
	}
}

func TestLLMThinkingLevel(t *testing.T) {
	cases := []struct {
		name string
		env  string
		want string
	}{
		{
			name: "low 原样返回",
			env:  "low",
			want: "low",
		},
		{
			name: "high 原样返回",
			env:  "high",
			want: "high",
		},
		{
			name: "max 原样返回",
			env:  "max",
			want: "max",
		},
		{
			name: "大写 LOW 大小写不敏感",
			env:  "LOW",
			want: "low",
		},
		{
			name: "空串视为未配置",
			env:  "",
			want: "",
		},
		{
			name: "medium 不受支持",
			env:  "medium",
			want: "",
		},
		{
			name: "disabled 不受支持",
			env:  "disabled",
			want: "",
		},
		{
			name: "含空白时去空白后归一化",
			env:  " low ",
			want: "low",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LLM_THINKING_LEVEL", tc.env)
			if got := llmThinkingLevel(); got != tc.want {
				t.Errorf("llmThinkingLevel()=%q, want %q(env=%q)", got, tc.want, tc.env)
			}
		})
	}
}
