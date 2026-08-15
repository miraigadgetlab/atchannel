package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kosero/atchannel/backend/internal/db"
	"github.com/kosero/atchannel/backend/internal/models"
)

type ReportRepo struct {
	q *db.Queries
}

func NewReportRepo(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{q: db.New(pool)}
}

func (r *ReportRepo) Create(ctx context.Context, targetType, targetID, reporterID, reason string) (*models.Report, error) {
	row, err := r.q.CreateReport(ctx, db.CreateReportParams{
		TargetType: targetType,
		TargetID:   targetID,
		ReporterID: reporterID,
		Reason:     reason,
	})
	if err != nil {
		return nil, err
	}
	return reportFromRow(row), nil
}

func (r *ReportRepo) GetByID(ctx context.Context, id string) (*models.Report, error) {
	row, err := r.q.GetReportByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return reportFromRow(row), nil
}

func (r *ReportRepo) UpdateStatus(ctx context.Context, id, status string) (*models.Report, error) {
	row, err := r.q.UpdateReportStatus(ctx, db.UpdateReportStatusParams{ID: id, Status: status})
	if err != nil {
		return nil, mapErr(err)
	}
	return reportFromRow(row), nil
}

func (r *ReportRepo) List(ctx context.Context, status string, limit, offset int) ([]models.Report, error) {
	rows, err := r.q.ListReports(ctx, db.ListReportsParams{
		Status: status,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	reports := make([]models.Report, 0, len(rows))
	for _, row := range rows {
		reports = append(reports, models.Report{
			ID:              row.ID,
			TargetType:      row.TargetType,
			TargetID:        row.TargetID,
			ReporterID:      row.ReporterID,
			Reason:          row.Reason,
			Status:          row.Status,
			CreatedAt:       row.CreatedAt,
			ReporterName:    row.ReporterName,
			TargetBoardSlug: row.TargetBoardSlug,
			TargetThreadID:  row.TargetThreadID,
			TargetBody:      row.TargetBody,
		})
	}
	return reports, nil
}

func reportFromRow(row db.Report) *models.Report {
	return &models.Report{
		ID:         row.ID,
		TargetType: row.TargetType,
		TargetID:   row.TargetID,
		ReporterID: row.ReporterID,
		Reason:     row.Reason,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
	}
}
