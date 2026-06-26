package service

import (
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

// withTimezone 临时把全局时区切到 tz,测试结束后恢复,避免污染其它用例。
func withTimezone(t *testing.T, tz string) {
	t.Helper()
	prev := timezone.Name()
	if prev == "Local" {
		prev = "UTC"
	}
	if err := timezone.Init(tz); err != nil {
		t.Fatalf("init timezone %q: %v", tz, err)
	}
	t.Cleanup(func() { _ = timezone.Init(prev) })
}

// 告警卡片的 Fired At 应按配置时区渲染,而非写死 UTC。
func TestBuildAlertCardFiredAtUsesConfiguredTimezone(t *testing.T) {
	withTimezone(t, "Asia/Shanghai")

	// 20:30 UTC == 次日 04:30 Asia/Shanghai,两个时刻不会混淆。
	fired := time.Date(2026, 6, 27, 20, 30, 0, 0, time.UTC)
	rule := &OpsAlertRule{Name: "错误率极高", Severity: "P0", MetricType: "error_rate"}
	event := &OpsAlertEvent{Severity: "P0", FiredAt: fired}

	flat := flattenAny(buildAlertCard(rule, event))

	if !strings.Contains(flat, "2026-06-28 04:30:00") {
		t.Errorf("expected fired at in Asia/Shanghai (2026-06-28 04:30:00), got: %s", flat)
	}
	if strings.Contains(flat, "2026-06-27 20:30:00") {
		t.Errorf("fired at should not be rendered in UTC, got: %s", flat)
	}
}

// 邮件模板变量 triggered_at 应按配置时区渲染(RFC3339 自带偏移)。
func TestOpsAlertEmailVariablesTriggeredAtUsesConfiguredTimezone(t *testing.T) {
	withTimezone(t, "Asia/Shanghai")

	fired := time.Date(2026, 6, 27, 20, 30, 0, 0, time.UTC)
	rule := &OpsAlertRule{Name: "错误率极高", Severity: "P0", MetricType: "error_rate"}
	event := &OpsAlertEvent{Status: "firing", FiredAt: fired}

	got := opsAlertEmailVariables(rule, event)["triggered_at"]
	want := fired.In(timezone.Location()).Format(time.RFC3339) // 2026-06-28T04:30:00+08:00

	if got != want {
		t.Errorf("triggered_at = %q, want %q", got, want)
	}
}
