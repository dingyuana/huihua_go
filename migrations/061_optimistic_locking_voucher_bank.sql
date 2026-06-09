-- 061_optimistic_locking_voucher_bank.sql: 加乐观锁 version 列
-- 用于"先读后写"场景的并发控制，跟 Phase 1 的悲观锁互补
-- - 悲观锁: 长事务 + 多表操作(核销/反核销)
-- - 乐观锁: 短事务 + 单行更新(凭证状态/银行流水分类)
--
-- 触发场景示例:
--   用户A打开凭证编辑(读到 version=3), 用户B同时修改并保存(version=4)
--   用户A提交时 UPDATE ... WHERE version=3 → 影响 0 行 → 提示"凭证已被他人修改,请刷新"

ALTER TABLE journal_entries
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 0;

ALTER TABLE bank_transactions
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN journal_entries.version IS '乐观锁版本号:每次 UPDATE 自增,WHERE version=$expected 检测并发更新冲突';
COMMENT ON COLUMN bank_transactions.version IS '乐观锁版本号:每次 UPDATE 自增,WHERE version=$expected 检测并发更新冲突';
