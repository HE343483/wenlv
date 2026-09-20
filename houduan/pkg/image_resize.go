package pkg

import (
	"bytes"
	"image"
	_ "image/gif" // 注册 GIF 解码器(仅用于识别格式,不参与压缩)
	"image/jpeg"
	"image/png"

	xdraw "golang.org/x/image/draw"
)

// 缩放宽度安全区间:避免异常参数导致内存暴涨或产出空图。
const (
	imageResizeMinWidth = 80
	imageResizeMaxWidth = 4000
)

// ResizeImage 按最大宽度等比压缩图片(JPEG/PNG),返回压缩后字节与 Content-Type。
// 以下情况返回 ok=false,调用方应沿用原始字节,保证行为与压缩前一致:
//   - 非 JPEG/PNG(动态 GIF、WebP 等):缩放会破坏动画或需要额外解码器;
//   - 解码失败或原图宽度未超过 maxWidth(不做放大);
//   - 压缩后体积反而更大。
//
// 输出格式与输入格式保持一致,因此调用方按原扩展名拼的 OSS key 仍然有效。
func ResizeImage(data []byte, maxWidth, quality int) ([]byte, string, bool) {
	if len(data) == 0 || maxWidth <= 0 {
		return nil, "", false
	}
	if maxWidth < imageResizeMinWidth {
		maxWidth = imageResizeMinWidth
	}
	if maxWidth > imageResizeMaxWidth {
		maxWidth = imageResizeMaxWidth
	}
	if quality <= 0 || quality > 100 {
		quality = 85
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width == 0 || cfg.Width <= maxWidth {
		return nil, "", false
	}
	if format != "jpeg" && format != "png" {
		return nil, "", false
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", false
	}

	height := cfg.Height * maxWidth / cfg.Width
	if height < 1 {
		height = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, maxWidth, height))
	// BiLinear 兼顾速度与画质(CatmullRom 在 6000px 级原图上过慢)
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)

	var buf bytes.Buffer
	contentType := "image/jpeg"
	if format == "png" {
		if err := png.Encode(&buf, dst); err != nil {
			return nil, "", false
		}
		contentType = "image/png"
	} else if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: quality}); err != nil {
		return nil, "", false
	}
	if buf.Len() >= len(data) {
		return nil, "", false
	}
	return buf.Bytes(), contentType, true
}
