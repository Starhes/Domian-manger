-- 添加DNS记录的新字段
ALTER TABLE dns_records 
ADD COLUMN IF NOT EXISTS weight INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS port INTEGER DEFAULT 0;

-- 更新comment字段长度
ALTER TABLE dns_records 
ALTER COLUMN comment TYPE VARCHAR(500);

-- 添加索引以提升查询性能
CREATE INDEX IF NOT EXISTS idx_dns_records_type_subdomain ON dns_records(type, subdomain);
CREATE INDEX IF NOT EXISTS idx_dns_records_domain_type ON dns_records(domain_id, type);

-- 更新现有记录的默认值
UPDATE dns_records SET weight = 0 WHERE weight IS NULL;
UPDATE dns_records SET port = 0 WHERE port IS NULL;