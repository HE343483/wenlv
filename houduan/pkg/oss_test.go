package pkg

import "testing"

const testWantHost = "https://wenlv-tdx.oss-cn-chengdu.aliyuncs.com"

// TestOssSignerEndpointNormalization 覆盖 Endpoint 带协议前缀 / 裸域名 / 带尾斜杠
// 三种写法下 ResolveURL 结果一致(修复前带前缀会拼出 https://wenlv-tdx.https://... )。
func TestOssSignerEndpointNormalization(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
	}{
		{"带 https 前缀", "https://oss-cn-chengdu.aliyuncs.com"},
		{"裸域名", "oss-cn-chengdu.aliyuncs.com"},
		{"带尾斜杠", "oss-cn-chengdu.aliyuncs.com/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewOssSigner(&OssConfig{
				Endpoint:  tc.endpoint,
				AccessKey: "ak",
				SecretKey: "sk",
				Bucket:    "wenlv-tdx",
			})
			if got, want := s.ResolveURL("scenic/1/g1.jpg"), testWantHost+"/scenic/1/g1.jpg"; got != want {
				t.Errorf("ResolveURL = %q, want %q", got, want)
			}
		})
	}
}

// TestOssSignerGeneratePolicyHost 校验上传 Policy 的 Host 与 ResolveURL 使用同一归一化 Endpoint。
func TestOssSignerGeneratePolicyHost(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
		cdnHost  string
		wantHost string
	}{
		{"带 https 前缀", "https://oss-cn-chengdu.aliyuncs.com", "", testWantHost},
		{"裸域名", "oss-cn-chengdu.aliyuncs.com", "", testWantHost},
		{"带尾斜杠", "oss-cn-chengdu.aliyuncs.com/", "", testWantHost},
		{"CdnHost 优先", "https://oss-cn-chengdu.aliyuncs.com", "https://img.example.com", "https://img.example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewOssSigner(&OssConfig{
				Endpoint:     tc.endpoint,
				AccessKey:    "ak",
				SecretKey:    "sk",
				Bucket:       "wenlv-tdx",
				PublicDomain: tc.cdnHost,
			})
			p, err := s.GeneratePolicy("scenic/1/g1.jpg", 10, nil)
			if err != nil {
				t.Fatalf("GeneratePolicy 出错: %v", err)
			}
			if p.Host != tc.wantHost {
				t.Errorf("Host = %q, want %q", p.Host, tc.wantHost)
			}
		})
	}
}

// TestOssSignerConfigured 校验配置完整性判断:四项齐全为 true,缺任一项为 false。
func TestOssSignerConfigured(t *testing.T) {
	base := OssConfig{
		Endpoint:  "oss-cn-chengdu.aliyuncs.com",
		AccessKey: "ak",
		SecretKey: "sk",
		Bucket:    "wenlv-tdx",
	}

	if !NewOssSigner(&base).Configured() {
		t.Errorf("四项齐全时 Configured() = false, want true")
	}

	cases := []struct {
		name   string
		mutate func(c *OssConfig)
	}{
		{"缺 Endpoint", func(c *OssConfig) { c.Endpoint = "" }},
		{"缺 AccessKey", func(c *OssConfig) { c.AccessKey = "" }},
		{"缺 SecretKey", func(c *OssConfig) { c.SecretKey = "" }},
		{"缺 Bucket", func(c *OssConfig) { c.Bucket = "" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.mutate(&cfg)
			if NewOssSigner(&cfg).Configured() {
				t.Errorf("缺项时 Configured() = true, want false")
			}
		})
	}
}
