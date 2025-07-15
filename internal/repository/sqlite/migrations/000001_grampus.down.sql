DROP INDEX IF EXISTS idx_trans_item_account_id;
DROP INDEX IF EXISTS idx_trans_item_trans_id;
DROP TABLE IF EXISTS trans_item;

DROP INDEX IF EXISTS idx_trans_book_id;
DROP TABLE IF EXISTS trans;

DROP INDEX IF EXISTS idx_account_name;
DROP INDEX IF EXISTS idx_account_book_id;
DROP TABLE IF EXISTS account;

DROP INDEX IF EXISTS idx_book_name;
DROP TABLE IF EXISTS book;

DROP INDEX IF EXISTS idx_price_comm_id;
DROP TABLE IF EXISTS price;

DROP TABLE IF EXISTS commodities;