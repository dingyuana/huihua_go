package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
	"huihua/finance/pkg/utils"
)

// PartyService handles party (customer/supplier) operations.
type PartyService struct {
	repo *repository.PartyRepository
}

// NewPartyService creates a new PartyService.
func NewPartyService(repo *repository.PartyRepository) *PartyService {
	return &PartyService{repo: repo}
}

// CreateParty creates a new party.
func (s *PartyService) CreateParty(ctx context.Context, tenantID uuid.UUID, p *model.Party) (*model.Party, error) {
	exists, err := s.repo.ExistsByNameAndType(ctx, tenantID, p.Name, p.PartyType)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("party with this name and type already exists")
	}
	return s.repo.Create(ctx, tenantID, p)
}

// List returns all active parties for a tenant.
func (s *PartyService) List(ctx context.Context, tenantID uuid.UUID, partyType string) ([]model.Party, error) {
	if partyType != "" {
		return s.repo.ListByType(ctx, tenantID, partyType)
	}
	return s.repo.List(ctx, tenantID)
}

// UpdateParty updates a party.
func (s *PartyService) UpdateParty(ctx context.Context, tenantID, id uuid.UUID, p *model.Party) error {
	return s.repo.Update(ctx, tenantID, id, p)
}

// DeleteParty soft-deletes a party.
func (s *PartyService) DeleteParty(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

// GetByID returns a party by ID.
func (s *PartyService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.Party, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

// ImportResult holds the result of an Excel import.
type ImportResult struct {
	SuccessCount int      `json:"success_count"`
	FailCount    int      `json:"fail_count"`
	FailRows     []int    `json:"fail_rows"`
	Errors       []string `json:"errors"`
}

// ImportExcel imports parties from an Excel file.
func (s *PartyService) ImportFromExcel(ctx context.Context, tenantID uuid.UUID, data []byte) (ImportResult, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return ImportResult{}, fmt.Errorf("open excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	sheetName := "Sheet1"
	if len(sheets) > 0 {
		sheetName = sheets[0]
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return ImportResult{}, fmt.Errorf("get rows: %w", err)
	}

	var parties []model.Party
	var failRows []int
	var errors []string

	// Expected columns (case-insensitive): party_type, name, tax_number, bank_name, bank_account,
	// contact_name, contact_phone, credit_limit, payment_days
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		rowNum := i + 2 // 1-based + header
		if len(row) < 2 {
			failRows = append(failRows, rowNum)
			errors = append(errors, fmt.Sprintf("row %d: insufficient columns", rowNum))
			continue
		}

		partyType := row[0]
		name := row[1]
		if partyType == "" || name == "" {
			failRows = append(failRows, rowNum)
			errors = append(errors, fmt.Sprintf("row %d: party_type and name are required", rowNum))
			continue
		}
		if partyType != "customer" && partyType != "supplier" && partyType != "both" {
			failRows = append(failRows, rowNum)
			errors = append(errors, fmt.Sprintf("row %d: invalid party_type '%s'", rowNum, partyType))
			continue
		}

		p := &model.Party{
			PartyType: partyType,
			Name:      name,
			IsActive:  true,
		}
		if len(row) > 2 && row[2] != "" {
			p.TaxNumber = utils.StrPtr(row[2])
		}
		if len(row) > 3 && row[3] != "" {
			p.BankName = utils.StrPtr(row[3])
		}
		if len(row) > 4 && row[4] != "" {
			p.BankAccount = utils.StrPtr(row[4])
		}
		if len(row) > 5 && row[5] != "" {
			p.ContactName = utils.StrPtr(row[5])
		}
		if len(row) > 6 && row[6] != "" {
			p.ContactPhone = utils.StrPtr(row[6])
		}
		if len(row) > 7 && row[7] != "" {
			if cl, err := decimal.NewFromString(row[7]); err == nil {
				p.CreditLimit = cl
			}
		}
		if len(row) > 8 && row[8] != "" {
			if pd, err := strconv.Atoi(row[8]); err == nil {
				p.PaymentDays = pd
			}
		}
		parties = append(parties, *p)
	}

	if len(parties) > 0 {
		if err := s.repo.CreateBatch(ctx, tenantID, parties); err != nil {
			return ImportResult{}, fmt.Errorf("batch insert: %w", err)
		}
	}

	return ImportResult{
		SuccessCount: len(parties),
		FailCount:    len(failRows),
		FailRows:     failRows,
		Errors:       errors,
	}, nil
}

