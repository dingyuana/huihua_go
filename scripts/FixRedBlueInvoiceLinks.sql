-- FixRedBlueInvoiceLinks.sql
-- 用途：扫描所有红字发票（is_return=true），从备注字段解析对应的蓝字发票号，
--       将蓝字发票标记为 is_reversed=true + status='reversed'
--
-- 支持的解析模式（与 extractBlueInvoiceNo() 保持一致）：
--   1. 对应蓝字发票号[：:]\s*(\d{8,20})
--   2. 对应正数发票号码[：:]\s*(\d{8,20})
--   3. 被红冲蓝字数电票号码[：:]\s*(\d{8,20})
--   4. 红冲发票(?:号[：:]?|[：:])?\s*(\d{8,20})
--   5. 蓝字发票(?:号[：:]?|[：:])?\s*(\d{8,20})
--
-- 使用方式：
--   psql -U huihua -d huihua_finance -f FixRedBlueInvoiceLinks.sql
--
-- 执行结果：输出所有被修复的发票对（红字发票号 → 蓝字发票号）

-- Step 1: 临时函数：正则匹配（PostgreSQL 9.x+ 支持正则）
DO $$
DECLARE
    r RECORD;
    blue_no TEXT;
    pattern TEXT;
    remark_text TEXT;
    matched BOOLEAN;
BEGIN
    RAISE NOTICE '=== 开始修复红字/蓝字发票关联 ===';

    -- 遍历所有红字发票
    FOR r IN
        SELECT id, invoice_no, remark, source_red_invoice_no, is_reversed
        FROM sales_invoices
        WHERE is_return = true
        ORDER BY posting_date DESC
    LOOP
        blue_no := NULL;
        matched := false;
        remark_text := COALESCE(r.remark, '');

        -- Priority 1: source_red_invoice_no 列（Excel 对应蓝字发票号 列）
        IF r.source_red_invoice_no IS NOT NULL AND r.source_red_invoice_no != '' THEN
            blue_no := r.source_red_invoice_no;
        END IF;

        -- Priority 2: 从备注字段解析
        IF blue_no IS NULL THEN
            -- 模式1: 对应蓝字发票号：xxx / 对应蓝字发票号: xxx
            IF remark_text ~ '对应蓝字发票号[：:]\s*(\d{8,20})' THEN
                blue_no := SUBSTRING(remark_text FROM '对应蓝字发票号[：:]\s*(\d{8,20})');
                matched := true;
            -- 模式2: 对应正数发票号码：xxx
            ELSIF remark_text ~ '对应正数发票号码[：:]\s*(\d{8,20})' THEN
                blue_no := SUBSTRING(remark_text FROM '对应正数发票号码[：:]\s*(\d{8,20})');
                matched := true;
            -- 模式3: 被红冲蓝字数电票号码：xxx
            ELSIF remark_text ~ '被红冲蓝字数电票号码[：:]\s*(\d{8,20})' THEN
                blue_no := SUBSTRING(remark_text FROM '被红冲蓝字数电票号码[：:]\s*(\d{8,20})');
                matched := true;
            -- 模式4: 红冲发票：xxx / 红冲发票号 xxx
            ELSIF remark_text ~ '红冲发票(?:号[：:]?|[：:])?\s*(\d{8,20})' THEN
                blue_no := SUBSTRING(remark_text FROM '红冲发票(?:号[：:]?|[：:])?\s*(\d{8,20})');
                matched := true;
            -- 模式5: 蓝字发票：xxx / 蓝字发票号 xxx
            ELSIF remark_text ~ '蓝字发票(?:号[：:]?|[：:])?\s*(\d{8,20})' THEN
                blue_no := SUBSTRING(remark_text FROM '蓝字发票(?:号[：:]?|[：:])?\s*(\d{8,20})');
                matched := true;
            END IF;
        END IF;

        -- Step 3: 如果找到蓝字发票号，标记蓝字发票
        IF blue_no IS NOT NULL AND blue_no != '' THEN
            UPDATE sales_invoices
            SET is_reversed = true,
                status = 'reversed'
            WHERE invoice_no = blue_no
              AND (is_reversed = false OR status != 'reversed');

            IF FOUND THEN
                RAISE NOTICE '✓ 红字发票 % → 标记蓝字发票 % 为已红冲', r.invoice_no, blue_no;
            ELSE
                RAISE NOTICE '⚠ 红字发票 % → 未在库中找到蓝字发票 %（可能跨租户或未导入）', r.invoice_no, blue_no;
            END IF;
        ELSE
            RAISE NOTICE '⚠ 红字发票 % 未找到对应蓝字发票号（备注: %)', r.invoice_no,
                         SUBSTRING(remark_text FROM 1 FOR 80);
        END IF;

    END LOOP;

    RAISE NOTICE '=== 修复完成 ===';
END $$;

-- 输出修复结果
SELECT '修复后的蓝字发票列表：' AS info;
SELECT invoice_no, status, is_reversed, source_red_invoice_no, remark
FROM sales_invoices
WHERE is_reversed = true
ORDER BY updated_at DESC;
