package service

import (
	"context"
	"fmt"
	"strings"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/engine"
	"cozy-insight/internal/model"
	"cozy-insight/internal/repository"
)

type DatasetService struct {
	repo         *repository.DatasetRepository
	dsRepo       *repository.DatasourceRepository
	rowPermRepo  *repository.RowPermissionRepository
	newConnector func(string) (engine.DatasourceConnector, error)
}

func NewDatasetService(repo *repository.DatasetRepository, dsRepo *repository.DatasourceRepository, rowPermRepo *repository.RowPermissionRepository) *DatasetService {
	return &DatasetService{
		repo:         repo,
		dsRepo:       dsRepo,
		rowPermRepo:  rowPermRepo,
		newConnector: engine.NewConnector,
	}
}

// SetConnectorFactory sets the connector factory (for testing).
func (s *DatasetService) SetConnectorFactory(f func(string) (engine.DatasourceConnector, error)) {
	s.newConnector = f
}

func (s *DatasetService) Create(ctx context.Context, req *dto.CreateDatasetRequest, userID uint64) (*model.Dataset, error) {
	ds := &model.Dataset{
		Name:         req.Name,
		DatasourceID: req.DatasourceID,
		DatabaseName: req.DatabaseName,
		TableName:    req.TableName,
		Type:         req.Type,
		Mode:         req.Mode,
		Status:       1,
		CreatedBy:    userID,
	}

	if err := s.repo.Create(ctx, ds); err != nil {
		return nil, err
	}
	return ds, nil
}

func (s *DatasetService) GetByID(ctx context.Context, id uint64) (*model.Dataset, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DatasetService) List(ctx context.Context) ([]model.Dataset, error) {
	return s.repo.List(ctx)
}

func (s *DatasetService) Update(ctx context.Context, id uint64, req *dto.UpdateDatasetRequest) error {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		ds.Name = req.Name
	}
	if req.DatasourceID != nil {
		ds.DatasourceID = *req.DatasourceID
	}
	if req.DatabaseName != "" {
		ds.DatabaseName = req.DatabaseName
	}
	if req.TableName != "" {
		ds.TableName = req.TableName
	}
	if req.Type != "" {
		ds.Type = req.Type
	}
	if req.Mode != nil {
		ds.Mode = *req.Mode
	}
	if req.Status != nil {
		ds.Status = *req.Status
	}

	return s.repo.Update(ctx, ds)
}

func (s *DatasetService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DatasetService) SyncFields(ctx context.Context, id uint64) error {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	datasource, err := s.dsRepo.FindByID(ctx, ds.DatasourceID)
	if err != nil {
		return fmt.Errorf("datasource not found: %w", err)
	}

	columns, err := s.getTableColumns(ctx, datasource, ds.DatabaseName, ds.TableName)
	if err != nil {
		return fmt.Errorf("get table columns failed: %w", err)
	}

	if err := s.repo.DeleteFields(ctx, id); err != nil {
		return err
	}

	var fields []model.DatasetField
	for _, col := range columns {
		deType := s.inferDeType(col.Type)
		fields = append(fields, model.DatasetField{
			DatasetID:  id,
			Name:       col.Name,
			Type:       col.Type,
			DeType:     deType,
			Length:     col.Length,
			Precision:  col.Precision,
			Scale:      col.Scale,
			OriginName: col.Name,
		})
	}

	if err := s.repo.CreateFields(ctx, fields); err != nil {
		return err
	}

	return nil
}

func (s *DatasetService) PreviewData(ctx context.Context, id uint64, limit uint64) (*dto.PreviewDataResponse, error) {
	ds, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	datasource, err := s.dsRepo.FindByID(ctx, ds.DatasourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	fields, err := s.repo.GetFields(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: 行级权限过滤（当前未接入真实用户属性系统）
	// rowFilter, _ := s.buildRowFilter(ctx, id, map[string]string{"dept_id": "1"})
	// if rowFilter != "" {
	//     // 将 rowFilter 注入 SQL WHERE
	// }

	data, err := s.queryTableData(ctx, datasource, ds.DatabaseName, ds.TableName, limit)
	if err != nil {
		return nil, fmt.Errorf("query table data failed: %w", err)
	}

	var fieldResp []dto.DatasetFieldResponse
	for _, f := range fields {
		fieldResp = append(fieldResp, dto.DatasetFieldResponse{
			ID:         f.ID,
			Name:       f.Name,
			Type:       f.Type,
			DeType:     f.DeType,
			Length:     f.Length,
			Precision:  f.Precision,
			Scale:      f.Scale,
			OriginName: f.OriginName,
		})
	}

	return &dto.PreviewDataResponse{
		Fields: fieldResp,
		Data:   data,
	}, nil
}

type columnInfo struct {
	Name      string
	Type      string
	Length    int
	Precision int
	Scale     int
}

func (s *DatasetService) getTableColumns(ctx context.Context, ds *model.Datasource, dbName, tableName string) ([]columnInfo, error) {
	conn, err := s.newConnector(ds.Type)
	if err != nil {
		return nil, fmt.Errorf("create connector failed: %w", err)
	}
	defer conn.Close()
	if err := conn.Connect(ds.Config); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	cols, err := conn.GetColumns(ctx, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}
	var result []columnInfo
	for _, c := range cols {
		result = append(result, columnInfo{
			Name:      c.Name,
			Type:      c.Type,
			Length:    c.Length,
			Precision: c.Precision,
			Scale:     c.Scale,
		})
	}
	return result, nil
}

func (s *DatasetService) queryTableData(ctx context.Context, ds *model.Datasource, dbName, tableName string, limit uint64) ([]map[string]interface{}, error) {
	conn, err := s.newConnector(ds.Type)
	if err != nil {
		return nil, fmt.Errorf("create connector failed: %w", err)
	}
	defer conn.Close()
	if err := conn.Connect(ds.Config); err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	var tableRef string
	if dbName != "" {
		tableRef = fmt.Sprintf("%s.%s", engine.QuoteIdentifier(dbName), engine.QuoteIdentifier(tableName))
	} else {
		tableRef = engine.QuoteIdentifier(tableName)
	}
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableRef, limit)
	return conn.Query(ctx, query)
}

func (s *DatasetService) inferDeType(sqlType string) int8 {
	sqlType = strings.ToUpper(sqlType)
	switch {
	case strings.Contains(sqlType, "INT"), strings.Contains(sqlType, "FLOAT"), strings.Contains(sqlType, "DOUBLE"), strings.Contains(sqlType, "DECIMAL"):
		return 2
	case strings.Contains(sqlType, "DATE"), strings.Contains(sqlType, "TIME"):
		return 1
	case strings.Contains(sqlType, "TEXT"), strings.Contains(sqlType, "VARCHAR"), strings.Contains(sqlType, "CHAR"):
		return 0
	default:
		return 4
	}
}

func (s *DatasetService) buildRowFilter(ctx context.Context, datasetID uint64, userAttrs map[string]string) ([]RowFilterCondition, error) {
	svc := NewRowPermissionService(s.rowPermRepo)
	return svc.BuildRowFilter(ctx, datasetID, userAttrs)
}

