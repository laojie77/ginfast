package models

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSysNoticeAddRequestValidateJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	body := []byte(`{"title":"999","content":"<p>8888</p>","type":3,"level":"H","publishNow":false,"targets":[{"targetType":1,"targetId":0,"includeChildren":false}]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/plugins/sysnotice/sysnotice/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req

	var payload SysNoticeAddRequest
	if err := payload.Validate(ctx); err != nil {
		t.Fatalf("validate returned error: %v", err)
	}
	if len(payload.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(payload.Targets))
	}
}
