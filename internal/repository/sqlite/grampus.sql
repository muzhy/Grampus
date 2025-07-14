PRAGMA foreign_keys = ON;

-- 商品，货币也是一种商品
CREATE TABLE IF NOT EXISTS commodities(
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,     -- 类型，用来表示货币，股票，基金，或者用户自定义的
    mnemonic TEXT NOT NULL, -- 商品名字简称，方便记忆
    fullname TEXT NOT NULL, -- 完整名称
    fraction INTEGER NOT NULL, -- 商品最小分割单位
    quote_flag INTEGER NOT NULL DEFAULT 0, -- 是否需要自动从外部获取价格信息
    quote_source TEXT DEFAULT NULL -- quote_flag为1时，使用quore_source获取
);

-- price 价格
-- 一样商品在不同的时间会有不同的价格，通过price
-- 来记录在不同商品在不同时间的价格
CREATE TABLE IF NOT EXISTS price(
    id TEXT PRIMARY KEY,
    comm_id TEXT NOT NULL, -- 商品id
    currency_id TEXT NOT NULL, -- 货币id，注意货币也是一种商品
    date TIMESTAMP NOT NULL,  
    source TEXT NOT NULL, -- 价格来源
    type TEXT NOT NULL, -- 价格类型，比如买入价，卖出价等
    value_num INTEGER NOT NULL, -- 与value_denom一起，表示价格的精确值
    value_denom INTEGER NOT NULL,
    FOREIGN KEY (comm_id) REFERENCES commodities(id),
    FOREIGN KEY (currency_id) REFERENCES commodities(id)
);

CREATE INDEX IF NOT EXISTS idx_price_comm_id ON price(comm_id);

-- 账簿
CREATE TABLE IF NOT EXISTS book (
    id TEXT PRIMARY KEY, -- UUID
    name TEXT NOT NULL UNIQUE,
    defalut_curr TEXT NOT NULL, -- 账簿使用的默认货币
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (defalut_curr) REFERENCES commodities(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_book_name ON book(name);

-- 科目，每个账簿都有自己的科目树
-- current_num 和 current_denom组合起来表示当前科目的余额，方便统计
CREATE TABLE IF NOT EXISTS account (
    id TEXT PRIMARY KEY, 
    book_id TEXT NOT NULL,  -- 所属账簿的id，方便进行检索
    parent_id TEXT DEFAULT NULL,  -- 通过parent_id关联父级科目，null表示顶级科目
    comm_id TEXT NOT NULL,      -- 每个科目都需要关联商品，货币，或者股票或其他
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    current_num INTEGER NOT NULL,
    current_denom INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (book_id) REFERENCES book(id),
    FOREIGN KEY (parent_id) REFERENCES account(id)
);

CREATE INDEX IF NOT EXISTS idx_account_book_id ON account(book_id);
CREATE INDEX IF NOT EXISTS idx_account_name ON account(name);

-- 交易，交易的具体细节在trans_item中
CREATE TABLE IF NOT EXISTS transaction (
    id TEXT PRIMARY KEY, 
    book_id TEXT NOT NULL, 
    transaction_date TIMESTAMP NOT NULL, 
    description TEXT,
    FOREIGN KEY (book_id) REFERENCES book(id)
);

CREATE INDEX IF NOT EXISTS idx_transaction_book_id ON transaction(book_id);

-- 交易的具体交易项，每笔交易至少包含两个交易项
-- amount_num 和 amount_denom结合起来表示精确的数字
-- 表示这笔交易有多少关联的account关联的商品
CREATE TABLE IF NOT EXISTS trans_item(
    id TEXT PRIMARY KEY,
    trans_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    desc TEXT DEFAULT NULL,
    amount_num INTEGER NOT NULL,
    amount_denom INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trans_id) REFERENCES transaction(id),
    FOREIGN KEY (account_id) REFERENCES account(id)
);

CREATE INDEX IF NOT EXISTS idx_trans_item_trans_id ON trans_item(trans_id);
CREATE INDEX IF NOT EXISTS idx_trans_item_account_id ON trans_item(account_id);
