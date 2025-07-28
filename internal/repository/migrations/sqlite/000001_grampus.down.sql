DROP INDEX IF EXISTS idx_entry_trans_id;
DROP INDEX IF EXISTS idx_entry_account_id;
DROP TABLE IF EXISTS entry;

DROP INDEX IF EXISTS idx_trans_ledger_id;
DROP TABLE IF EXISTS trans;

DROP INDEX IF EXISTS idx_account_name;
DROP INDEX IF EXISTS idx_account_ledger_id;
DROP TABLE IF EXISTS account;

DROP INDEX IF EXISTS idx_ledger_name;
DROP TABLE IF EXISTS ledger;

DROP INDEX IF EXISTS idx_price_goods_id;
DROP TABLE IF EXISTS price;

DROP TABLE IF EXISTS goods;

DROP TABLE IF EXISTS account_type;
