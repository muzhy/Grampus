PRAGMA foreign_keys = ON;

-- 主键ID 统一使用UUID，在Sqlite中由程序负责生成
-- 由于Sqlite中不支持decimal，为表示精确的数字
-- 数据采用分数形式存储，分为分子和分母两部分，
-- 通常分母为10的整数倍

-- 账户类型表
CREATE TABLE IF NOT EXISTS account_type (
    type TEXT PRIMARY KEY
);

INSERT INTO account_type (type) VALUES
    ('asset'),              -- 资产
    ('liability'),          -- 负债
    ('equity'),             -- 权益
    ('revenue'),            -- 收入
    ('expense');            -- 支出


-- 商品，货币也视为一种商品
CREATE TABLE IF NOT EXISTS goods(
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,                         -- 类型，用来表示货币，股票，基金
    mnemonic TEXT NOT NULL,                     -- 商品名字简称，方便记忆
    fullname TEXT NOT NULL,                     -- 完整名称
    fraction INTEGER NOT NULL,                  -- 商品最小分割单位
    quote_flag INTEGER NOT NULL DEFAULT 0,      -- 是否需要自动从外部获取价格信息
    quote_source TEXT DEFAULT NULL             -- quote_flag为1时，使用quore_source获取
);
-- TODO 后续需要考虑多用户的场景，商品可以系统内置，也可以用户添加
-- 需要指出不同用户之间的商品的隔离和共享

-- 初始内置的货币
INSERT INTO goods (id, type, mnemonic, fullname, fraction, quote_flag, quote_source)
VALUES
    ('CNY', 'currency', 'CNY', 'Chinese Yuan', 100, 0, NULL),
    ('USD', 'currency', 'USD', 'US Dollar', 100, 0, NULL),
    ('EUR', 'currency', 'EUR', 'Euro', 100, 0, NULL),
    ('JPY', 'currency', 'JPY', 'Japanese Yen', 100, 0, NULL),
    ('GBP', 'currency', 'GBP', 'British Pound', 100, 0, NULL),
    ('AUD', 'currency', 'AUD', 'Australian Dollar', 100, 0, NULL),
    ('CAD', 'currency', 'CAD', 'Canadian Dollar', 100, 0, NULL),
    ('CHF', 'currency', 'CHF', 'Swiss Franc', 100, 0, NULL),
    ('HKD', 'currency', 'HKD', 'Hong Kong Dollar', 100, 0, NULL),
    ('SGD', 'currency', 'SGD', 'Singapore Dollar', 100, 0, NULL);

-- price 价格
-- 一样商品在不同的时间会有不同的价格，通过price
-- 来记录在不同商品在不同时间的价格
CREATE TABLE IF NOT EXISTS price(
    id TEXT PRIMARY KEY,
    goods_id TEXT NOT NULL,                     -- 商品id
    currency_id TEXT NOT NULL,                  -- 货币id，注意货币也是一种商品
    date TIMESTAMP NOT NULL,                    -- 当前价格产生的日期
    source TEXT NOT NULL,                       -- 价格来源
    type TEXT NOT NULL,                         -- 价格类型，比如买入价，卖出价等
    value_num INTEGER NOT NULL,                 -- 与value_denom一起，表示价格的精确值
    value_denom INTEGER NOT NULL,
    FOREIGN KEY (goods_id) REFERENCES goods(id),
    FOREIGN KEY (currency_id) REFERENCES goods(id)
);

CREATE INDEX IF NOT EXISTS idx_price_goods_id ON price(goods_id);

-- 账簿
CREATE TABLE IF NOT EXISTS ledger (
    id TEXT PRIMARY KEY, 
    name TEXT NOT NULL UNIQUE,  
    defalut_curr TEXT NOT NULL,         -- 账簿使用的默认货币
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (defalut_curr) REFERENCES goods(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ledger_name ON ledger(name);

-- 科目，每个账簿都有自己的科目树
-- current_num 和 current_denom组合起来表示当前科目的余额，方便统计
CREATE TABLE IF NOT EXISTS account (
    id TEXT PRIMARY KEY, 
    ledger_id TEXT NOT NULL,            -- 所属账簿的id，方便进行检索
    parent_id TEXT DEFAULT NULL,        -- 通过parent_id关联父级科目，null表示顶级科目
    level int NOT NULL,                 -- 科目的层级，为防止层级过深，也为了优化加载
    goods_id TEXT NOT NULL,             -- 每个科目都需要关联商品，货币，或者股票或其他
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,                 -- 路径,
    type TEXT NOT NULL,                 -- 每个科目必须属于5种account_type之一
    current_num INTEGER NOT NULL,
    current_denom INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (ledger_id) REFERENCES ledger(id),
    FOREIGN KEY (type) REFERENCES account_type(type),
    UNIQUE (parent_id, name)  -- 相同父节点的下的子节点的名称必须唯一
);

CREATE INDEX IF NOT EXISTS idx_account_ledger_id ON account(ledger_id);
CREATE INDEX IF NOT EXISTS idx_account_name ON account(name);
CREATE INDEX IF NOT EXISTS idx_account_parent_id ON account(parent_id);

-- 交易
-- 每笔交易至少需要包含两个交易项
-- 基本原则: 有借必有贷,借贷必相等
CREATE TABLE IF NOT EXISTS trans (
    id TEXT PRIMARY KEY, 
    ledger_id TEXT NOT NULL, 
    trans_date TIMESTAMP NOT NULL, 
    description TEXT,
    FOREIGN KEY (ledger_id) REFERENCES ledger(id)
);

CREATE INDEX IF NOT EXISTS idx_trans_ledger_id ON trans(ledger_id);

-- 交易的具体交易项，每笔交易至少包含两个交易项
-- amount_num 和 amount_denom结合起来表示精确的数字
CREATE TABLE IF NOT EXISTS entry(
    id TEXT PRIMARY KEY,
    trans_id TEXT NOT NULL,
    account_id TEXT NOT NULL,
    desc TEXT DEFAULT NULL,
    amount_num INTEGER NOT NULL,
    amount_denom INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trans_id) REFERENCES trans(id),
    FOREIGN KEY (account_id) REFERENCES account(id)
);

CREATE INDEX IF NOT EXISTS idx_entry_trans_id ON entry(trans_id);
CREATE INDEX IF NOT EXISTS idx_entry_account_id ON entry(account_id);
