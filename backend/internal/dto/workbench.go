package dto

import "time"

// WorkbenchStatsResponse 工作台资源统计响应
type WorkbenchStatsResponse struct {
	DatasourceCount int64 `json:"datasourceCount"`
	DatasetCount    int64 `json:"datasetCount"`
	ChartCount      int64 `json:"chartCount"`
	DashboardCount  int64 `json:"dashboardCount"`
	ScreenCount     int64 `json:"screenCount"`
}

// RecentViewItem 最近访问项（联表查询结果）
type RecentViewItem struct {
	ID        uint64    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Type      string    `db:"type" json:"type"`
	VisitedAt time.Time `db:"visited_at" json:"visitedAt"`
}

// FavoriteItem 收藏项（联表查询结果）
type FavoriteItem struct {
	ID        uint64    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Type      string    `db:"type" json:"type"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// RecordVisitRequest 记录访问请求
type RecordVisitRequest struct {
	ResourceType string `json:"resourceType" binding:"required,oneof=dashboard screen"`
	ResourceID   uint64 `json:"resourceId" binding:"required"`
}

// AddFavoriteRequest 添加收藏请求
type AddFavoriteRequest struct {
	ResourceType string `json:"resourceType" binding:"required"`
	ResourceID   uint64 `json:"resourceId" binding:"required"`
}
