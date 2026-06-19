package model

import "time"

// ScheduleTask 定时任务配置。
//
// 调度规则由 CronExpr 指定(如 "0 9 * * *"),Type 表示投递类型
// (email_report / snapshot / data_sync 等),Service 通过
// ScheduleService.RegisterTaskType 注入实际 handler。
//
// Config 是 JSON 字符串,业务侧可自由扩展(例如邮件收件人、目标报表 ID)。
// 每次执行的结果追加到 Config 的 last_runs 列表(由 service.recordTaskRun
// 维护,最近 10 条),用于运维审计。
type ScheduleTask struct {
	ID         uint64    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Type       string    `db:"type" json:"type"`
	CronExpr   string    `db:"cron_expr" json:"cronExpr"`
	Config     string    `db:"config" json:"config"`
	Enabled    bool      `db:"enabled" json:"enabled"`
	Status     string    `db:"status" json:"status"`
	CreatedBy  uint64    `db:"created_by" json:"createdBy"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	UpdateTime time.Time `db:"update_time" json:"updateTime"`
}
