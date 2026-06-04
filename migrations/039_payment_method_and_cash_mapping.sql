-- 039: Add payment_method to payment_entries + cash scenario mappings
-- payment_method: bank/cash/wechat/alipay/other — determines bank-side account (1001 vs 1002)

-- Add payment_method column
ALTER TABLE payment_entries
  ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20);

-- Add cash scenario mappings for receipt and payment
-- When payment_method='cash', the bank side should use 1001 (库存现金) instead of 1002
INSERT INTO bus_doc_mapping (id, tenant_id, doc_type, condition_key, condition_label, debit_subject_code, debit_subject_name, credit_subject_code, credit_subject_name, sort_order)
VALUES
  -- 收款单-现金：借 1001 库存现金，贷 1122 应收账款
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'receipt', 'cash', '现金收款', '1001', '库存现金', '1122', '应收账款', 1),
  -- 付款单-现金：借 2202 应付账款，贷 1001 库存现金
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'payment', 'cash', '现金付款', '2202', '应付账款', '1001', '库存现金', 1)
ON CONFLICT DO NOTHING;
