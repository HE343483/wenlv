package handler

import (
	"github.com/gin-gonic/gin"

	"wenlv-backend/middleware"
	"wenlv-backend/pkg"
	"wenlv-backend/service"
)

// UploadHandler 图片直传接口。
type UploadHandler struct {
	svc *service.UploadService
}

// NewUploadHandler 构造上传处理器。
func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

type uploadPolicyReq struct {
	FileName string `json:"file_name"`
}

// Policy 返回 OSS 直传签名。
func (h *UploadHandler) Policy(c *gin.Context) {
	var req uploadPolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.BadRequest(c, "参数错误")
		return
	}
	uid := middleware.GetUID(c)
	policy, err := h.svc.Policy(uid, req.FileName)
	if err != nil {
		pkg.ServerErrorWithErr(c, err, "生成上传签名失败")
		return
	}
	pkg.OK(c, gin.H{
		"policy":     policy,
		"object_url": h.svc.ResolveURL(policy.Key),
		"expires_in": 15 * 60,
	})
}
