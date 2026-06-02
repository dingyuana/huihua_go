BEGIN;

CREATE OR REPLACE FUNCTION extract_counterparty_name(description TEXT) RETURNS TEXT AS $$
DECLARE
  result TEXT;
BEGIN
  result := (regexp_match(description, '国家税务总局[\u4e00-\u9fa5]{0,15}税务局|[\u4e00-\u9fa5]{2,20}税务局'))[1];
  IF result IS NOT NULL THEN RETURN trim(result); END IF;

  result := (regexp_match(description, '[\u4e00-\u9fa5]{2,20}(社保局|公积金中心|社保中心|海关)'))[0];
  IF result IS NOT NULL THEN RETURN trim(result); END IF;

  result := (regexp_match(description, '[\u4e00-\u9fa5]{2,30}(有限公司|股份有限公司|集团|有限责任公司|股份公司|总公司|分公司|子公司|集团公司)'))[0];
  IF result IS NOT NULL THEN RETURN trim(result); END IF;

  result := (regexp_match(description, '[\u4e00-\u9fa5]{4,20}(公司|厂|店|商行|银行|事务所|医院|学校|中心)'))[0];
  IF result IS NOT NULL THEN RETURN trim(result); END IF;

  RETURN NULL;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

WITH polluted AS (
  SELECT id, description, direction
  FROM bank_transactions
  WHERE counterparty_name = '10086' OR counterparty_name LIKE '10086%'
),
extracted AS (
  SELECT
    p.id,
    p.direction,
    extract_counterparty_name(p.description) AS new_counterparty
  FROM polluted p
),
upd AS (
  UPDATE bank_transactions bt
  SET counterparty_name = e.new_counterparty
  FROM extracted e
  WHERE bt.id = e.id
    AND bt.counterparty_name IS DISTINCT FROM e.new_counterparty
  RETURNING bt.id
)
INSERT INTO parties (id, tenant_id, party_type, name, is_active, created_at, updated_at)
SELECT
  gen_random_uuid(),
  '00000000-0000-0000-0000-000000000001',
  CASE WHEN e.direction = 'in' THEN 'customer' WHEN e.direction = 'out' THEN 'supplier' ELSE 'both' END,
  e.new_counterparty,
  true,
  now(),
  now()
FROM (SELECT DISTINCT direction, new_counterparty FROM extracted WHERE new_counterparty IS NOT NULL AND new_counterparty <> '') e
WHERE NOT EXISTS (
  SELECT 1 FROM parties p
  WHERE p.tenant_id = '00000000-0000-0000-0000-000000000001'
    AND p.name = e.new_counterparty
    AND p.party_type = (CASE WHEN e.direction = 'in' THEN 'customer' WHEN e.direction = 'out' THEN 'supplier' ELSE 'both' END)
);

DROP FUNCTION extract_counterparty_name(TEXT);

COMMIT;

SELECT classification, counterparty_name, COUNT(*) AS cnt
FROM bank_transactions
WHERE id IN (
  SELECT id FROM bank_transactions
  WHERE description LIKE '%山东恺拓%' OR description LIKE '%手续费%' OR description LIKE '%OBSS%'
)
GROUP BY classification, counterparty_name
ORDER BY classification, counterparty_name;
