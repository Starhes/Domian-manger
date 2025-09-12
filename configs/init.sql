-- 域名管理系统初始化SQL脚本

-- 创建管理员用户（密码：admin123，请在首次登录后立即修改）
INSERT INTO users (email, password, nickname, is_active, is_admin, dns_record_quota, status, created_at, updated_at)
VALUES (
    'admin@example.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash of 'admin123'
    '系统管理员',
    true,
    true,
    1000,
    'normal',
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 创建示例域名
INSERT INTO domains (name, domain_type, is_active, description, created_at, updated_at)
VALUES 
    ('example.com', '二级域名', true, '示例域名', NOW(), NOW()),
    ('test.org', '二级域名', true, '测试域名', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 创建示例DNS服务商配置
INSERT INTO dns_providers (name, type, config, is_active, description, sort_order, created_at, updated_at)
VALUES (
    'DNSPod',
    'dnspod',
    '{"api_token": "请在管理后台配置实际的API Token"}',
    false,
    '腾讯云DNSPod服务商',
    1,
    NOW(),
    NOW()
) ON CONFLICT (name) DO NOTHING;

-- 创建示例SMTP配置
INSERT INTO smtp_configs (name, host, port, username, password, from_email, from_name, is_active, is_default, use_tls, description, created_at, updated_at)
VALUES (
    '默认SMTP配置',
    'smtp.gmail.com',
    587,
    'your_email@gmail.com',
    '请在管理后台配置实际的密码',
    'noreply@example.com',
    'Domain MAX',
    false,
    true,
    true,
    '默认邮件发送配置，请在管理后台修改为实际配置',
    NOW(),
    NOW()
) ON CONFLICT (name) DO NOTHING;