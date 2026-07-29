package service

import (
	"fmt"
	"log"
	"sync"

	"go-admin/common/auth/model"
	"go-admin/common/auth/repository"

	"gorm.io/gorm"
)

const operationLogBufferSize = 1000

// AsyncOperationLogger 异步操作日志记录器
type AsyncOperationLogger struct {
	db       *gorm.DB
	repo     repository.OperationLogRepository
	ch       chan *model.OperationLog
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewAsyncOperationLogger 创建异步日志记录器
func NewAsyncOperationLogger(db *gorm.DB, repo repository.OperationLogRepository) *AsyncOperationLogger {
	return &AsyncOperationLogger{
		db:     db,
		repo:   repo,
		ch:     make(chan *model.OperationLog, operationLogBufferSize),
		stopCh: make(chan struct{}),
	}
}

// Start 启动 worker goroutine
func (l *AsyncOperationLogger) Start() {
	l.wg.Add(1)
	go l.worker()
}

// Stop 停止 worker，等待处理完剩余日志
func (l *AsyncOperationLogger) Stop() {
	close(l.stopCh)
	l.wg.Wait()
}

// highRiskActions 高风险操作 action 白名单
var highRiskActions = map[string]bool{
	"delete_tenant":          true,
	"delete_role":            true,
	"delete_user":            true,
	"force_offline":          true,
	"assign_resources":       true,
	"assign_apis":            true,
	"cascade_trim_resources": true,
	"cascade_trim_apis":      true,
}

// Log 非阻塞写入日志，channel 满时降级打印标准日志
func (l *AsyncOperationLogger) Log(entry *model.OperationLog) {
	// CR-8: 自动填充 risk_level（如果未设置）
	if entry.RiskLevel == "" {
		if highRiskActions[entry.Action] {
			entry.RiskLevel = "HIGH"
		} else {
			entry.RiskLevel = "LOW"
		}
	}

	select {
	case l.ch <- entry:
	default:
		log.Printf("[operation-log] channel 已满，降级输出: module=%s action=%s user_id=%d target_type=%s target_id=%s summary=%s",
			entry.Module, entry.Action, entry.UserID, entry.TargetType, entry.TargetID, entry.Summary)
	}
}

// LogSimple 简化的日志记录方法，兼容 OperationLogger 接口调用模式
func (l *AsyncOperationLogger) LogSimple(operatorID int64, action string, targetType string, targetID int64, detail string) {
	entry := &model.OperationLog{
		UserID:     operatorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   formatInt64(targetID),
		Summary:    detail,
	}
	l.Log(entry)
}

// worker 从 channel 读取条目写入 DB
func (l *AsyncOperationLogger) worker() {
	defer l.wg.Done()
	for {
		select {
		case entry, ok := <-l.ch:
			if !ok {
				return
			}
			if err := l.repo.Create(l.db, entry); err != nil {
				log.Printf("[operation-log] 写入失败: %v, entry: module=%s action=%s",
					err, entry.Module, entry.Action)
			}
		case <-l.stopCh:
			// 排空 channel 中剩余条目
			for {
				select {
				case entry, ok := <-l.ch:
					if !ok {
						return
					}
					if err := l.repo.Create(l.db, entry); err != nil {
						log.Printf("[operation-log] 写入失败(draining): %v", err)
					}
				default:
					return
				}
			}
		}
	}
}

// formatInt64 将 int64 转为字符串
func formatInt64(v int64) string {
	return fmt.Sprintf("%d", v)
}

// OperationLogListResult 操作日志分页结果
type OperationLogListResult struct {
	List     []model.OperationLog `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// OperationLogQueryService 操作日志查询服务
type OperationLogQueryService struct {
	db   *gorm.DB
	repo repository.OperationLogRepository
}

// NewOperationLogQueryService 创建查询服务
func NewOperationLogQueryService(db *gorm.DB, repo repository.OperationLogRepository) *OperationLogQueryService {
	return &OperationLogQueryService{db: db, repo: repo}
}

// ListLogs 分页查询操作日志
func (s *OperationLogQueryService) ListLogs(params repository.OperationLogListParams) (*OperationLogListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	list, total, err := s.repo.List(s.db, params)
	if err != nil {
		return nil, err
	}

	return &OperationLogListResult{
		List:     list,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}
