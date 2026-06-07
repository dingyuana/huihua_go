-- Migration 2026060806: payment_entries 加币种 + 汇率字段
--
-- 业务背景:
--   V1.0 报告 P0 缺失：payment_entries 缺少币种字段。Go 版模型一直隐式 CNY，
--   后续跨境/外币付款场景（USD/EUR/HKD 等）无法表达。
--   本次新增 currency + exchange_rate，为 V2 远期多币种扩展留位。
--
-- V1/V2 范围:
--   V1.1: 仅 model 留位 + DB 默认 CNY/1.0，service/handler 不强制多币种。
--   V2 远期: 多币种付款、汇率换算、汇兑损益。
--
-- 字段语义:
--   currency      — 币种代码（ISO 4217 三位字母：CNY/USD/EUR/HKD）
--   exchange_rate — 折算汇率（默认 1.0，后续 V2 按日维护）
--
-- 数据兼容性:
--   现有行 currency 默认 'CNY'、exchange_rate 默认 1.0，行为与原隐式 CNY 完全一致。

ALTER TABLE payment_entries
  ADD COLUMN IF NOT EXISTS currency      VARCHAR(3)    NOT NULL DEFAULT 'CNY';
ALTER TABLE payment_entries
  ADD COLUMN IF NOT EXISTS exchange_rate NUMERIC(18,6) NOT NULL DEFAULT 1.0;

COMMENT ON COLUMN payment_entries.currency IS
  '币种代码（CNY/USD/EUR/HKD），V2 远期扩展';
COMMENT ON COLUMN payment_entries.exchange_rate IS
  '折算汇率（V1.1 默认 1.0，V2 远期按日维护）';

-- 回滚（如需撤销）
-- ALTER TABLE payment_entries DROP COLUMN IF EXISTS exchange_rate;
-- ALTER TABLE payment_entries DROP COLUMN IF EXISTS currency;
