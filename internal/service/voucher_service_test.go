package service

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

func newTestVoucherService(t *testing.T) *VoucherService {
	t.Helper()
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("SKIP_INTEGRATION=1")
	}
	pool, err := pgxpool.New(testCtx, "postgres://huihua:hfpwd@127.0.0.1:5432/huihua_finance?sslmode=disable")
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	return NewVoucherService(
		repository.NewJournalRepository(pool),
		NewVoucherTemplateService(
			repository.NewVoucherTemplateRepository(pool),
			repository.NewAccountRepository(pool),
		),
	)
}

func createAccountForTest(t *testing.T, acctRepo *repository.AccountRepository, code, name string) uuid.UUID {
	t.Helper()
	acct := &model.Account{
		Code:      code + "-" + uuid.New().String()[:8],
		Name:      name,
		CompanyID: testTenantID,
		TenantID:  testTenantID,
		Currency:  "CNY",
		IsActive:  true,
		Lft:       100,
		Rgt:       101,
		IsGroup:   false,
	}
	created, err := acctRepo.Create(testCtx, testTenantID, acct)
	if err != nil {
		t.Fatalf("create account %s: %v", code, err)
	}
	return created.ID
}

func TestVoucherService_CreateVoucher(t *testing.T) {
	skipIfNoDB(t)
	pool := testPool(t)
	svc := newTestVoucherService(t)
	acctRepo := repository.NewAccountRepository(pool)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	companyID := testTenantID
	voucherType := "journal_entry"
	remark := "Test voucher from integration test"

	acctID1 := createAccountForTest(t, acctRepo, "VCTST1", "Voucher Test Debit")
	acctID2 := createAccountForTest(t, acctRepo, "VCTST2", "Voucher Test Credit")

	voucher := &CreateVoucherRequest{
		CompanyID:   companyID,
		VoucherType: &voucherType,
		PostingDate: time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
		Remark:      &remark,
		Lines: []VoucherLineRequest{
			{AccountID: acctID1, Debit: "1000.00", Credit: "0.00"},
			{AccountID: acctID2, Debit: "0.00", Credit: "1000.00"},
		},
	}

	je, err := svc.CreateVoucher(testCtx, testTenantID, userID, voucher)
	if err != nil {
		t.Fatalf("CreateVoucher failed: %v", err)
	}
	t.Logf("Created voucher: id=%s no=%s status=%d", je.ID.String(), je.VoucherNo, je.DocStatus)
}

func TestVoucherService_UpdateVoucher(t *testing.T) {
	skipIfNoDB(t)
	pool := testPool(t)
	svc := newTestVoucherService(t)
	acctRepo := repository.NewAccountRepository(pool)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	companyID := testTenantID
	voucherType := "journal_entry"
	remark := "Test update"

	acctID1 := createAccountForTest(t, acctRepo, "VCTST3", "Voucher Test Debit 2")
	acctID2 := createAccountForTest(t, acctRepo, "VCTST4", "Voucher Test Credit 2")

	voucher := &CreateVoucherRequest{
		CompanyID:   companyID,
		VoucherType: &voucherType,
		PostingDate: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		Remark:      &remark,
		Lines: []VoucherLineRequest{
			{AccountID: acctID1, Debit: "500.00", Credit: "0.00"},
			{AccountID: acctID2, Debit: "0.00", Credit: "500.00"},
		},
	}

	je, err := svc.CreateVoucher(testCtx, testTenantID, userID, voucher)
	if err != nil {
		t.Fatalf("CreateVoucher failed: %v", err)
	}
	t.Logf("Created voucher: %s", je.VoucherNo)

	newRemark := "Updated remark"
	err = svc.UpdateVoucher(testCtx, testTenantID, je.ID, userID, &UpdateVoucherRequest{
		Remark: &newRemark,
	})
	if err != nil {
		t.Fatalf("UpdateVoucher failed: %v", err)
	}
	t.Log("Update succeeded")
}
