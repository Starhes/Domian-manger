-- 初始化数据库脚本

-- 创建管理员用户 (密码: admin123)
-- 注意：生产环境请修改密码
INSERT INTO users (email, password, is_active, is_admin, created_at, updated_at) 
VALUES (
    'admin@example.com', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- admin123
    true, 
    true, 
    NOW(), 
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 插入示例域名
INSERT INTO domains (name, is_active, created_at, updated_at) 
VALUES 
    ('example.com', true, NOW(), NOW()),
    ('test.com', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 插入示例DNS服务商配置
INSERT INTO dns_providers (name, type, config, is_active, created_at, updated_at) 
VALUES (
    'DNSPod默认',
    'dnspod',
    '{"token": "your_dnspod_token_here"}',
    false,  -- 默认禁用，需要管理员配置正确的token后启用
    NOW(),
    NOW()
) ON CONFLICT DO NOTHING;
