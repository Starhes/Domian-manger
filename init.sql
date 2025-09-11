-- 初始化数据库脚本

-- 创建管理员用户 (密码: admin123)
-- 注意：生产环境请修改密码
INSERT INTO users (email, password, nickname, is_active, is_admin, status, dns_record_quota, login_count, created_at, updated_at) 
VALUES (
    'admin@example.com', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    '系统管理员',
    true, 
    true,
    'normal',
    0,  -- 管理员不限制配额
    0,
    NOW(), 
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 插入示例域名
INSERT INTO domains (name, domain_type, description, is_active, created_at, updated_at) 
VALUES 
    ('example.com', '二级域名', '示例主域名，用于测试和演示', true, NOW(), NOW()),
    ('test.com', '二级域名', '测试域名，用于开发环境', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 插入示例DNS服务商配置
INSERT INTO dns_providers (name, type, config, description, is_active, sort_order, created_at, updated_at) 
VALUES (
    'DNSPod默认',
    'dnspod',
    '{"token": "your_dnspod_token_here"}',
    'DNSPod传统API接口，格式为ID,Token',
    false,  -- 默认禁用，需要管理员配置正确的token后启用
    1,
    NOW(),
    NOW()
) ON CONFLICT (name) DO NOTHING;
