# 网文大纲续写工具数据库设计

## 数据库选择
推荐使用MySQL或PostgreSQL关系型数据库，以确保数据完整性和支持复杂查询。项目同时支持SQLite作为轻量级选项。

## 表结构设计

### 1. 用户表 (users)

```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码(哈希值)',
    display_name VARCHAR(50) COMMENT '显示名称',
    email VARCHAR(100) UNIQUE COMMENT '电子邮箱',
    role INT NOT NULL DEFAULT 1 COMMENT '角色(1:普通用户，2:管理员)', 
    status INT NOT NULL DEFAULT 1 COMMENT '状态(1:启用，0:禁用)',
    token VARCHAR(255) COMMENT '身份令牌/Token余额',
    github_id VARCHAR(50) COMMENT 'GitHub ID',
    wechat_id VARCHAR(50) COMMENT '微信ID',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_token (token),
    INDEX idx_github_id (github_id),
    INDEX idx_wechat_id (wechat_id)
);
```

### 2. 推荐码表 (referrals)

```sql
CREATE TABLE referrals (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL UNIQUE COMMENT '用户ID',
    code VARCHAR(20) NOT NULL UNIQUE COMMENT '推荐码',
    total_used INT NOT NULL DEFAULT 0 COMMENT '使用次数',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_user_id (user_id)
);
```

### 3. 推荐使用记录表 (referral_uses)

```sql
CREATE TABLE referral_uses (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    referrer_id INT UNSIGNED NOT NULL COMMENT '推荐人ID',
    user_id INT UNSIGNED NOT NULL UNIQUE COMMENT '被推荐人ID',
    referral_code VARCHAR(20) NOT NULL COMMENT '使用的推荐码',
    tokens_rewarded INT NOT NULL COMMENT '奖励的token数量',
    used_at TIMESTAMP NOT NULL COMMENT '使用时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_referrer_id (referrer_id),
    INDEX idx_user_id (user_id)
);
```

### 4. Token记录表 (token_records)

```sql
CREATE TABLE token_records (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    amount INT NOT NULL COMMENT '金额变动(正值为增加，负值为消费)',
    balance INT NOT NULL COMMENT '变动后余额',
    record_type INT NOT NULL COMMENT '记录类型(1:套餐赠送,2:推荐奖励,3:续写消费,4:充值)',
    related_id INT COMMENT '相关记录ID',
    description VARCHAR(255) COMMENT '描述',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_record_type (record_type)
);
```

### 5. 套餐表 (packages)

```sql
CREATE TABLE packages (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL COMMENT '套餐名称',
    description VARCHAR(255) COMMENT '套餐描述',
    price DECIMAL(10,2) NOT NULL COMMENT '价格',
    monthly_tokens INT NOT NULL COMMENT '每月赠送token数量',
    duration VARCHAR(20) NOT NULL COMMENT '有效期类型(monthly,yearly,permanent)',
    features TEXT COMMENT '功能列表(JSON字符串)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### 6. 订阅表 (subscriptions)

```sql
CREATE TABLE subscriptions (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
    package_id INT UNSIGNED NOT NULL COMMENT '套餐ID',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态(active,expired,cancelled)',
    start_date TIMESTAMP NOT NULL COMMENT '开始日期',
    expiry_date TIMESTAMP NOT NULL COMMENT '过期日期',
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否自动续费',
    next_renewal TIMESTAMP COMMENT '下次续费时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_package_id (package_id)
);
```

### 7. Token分发记录表 (token_distributions)

```sql
CREATE TABLE token_distributions (
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL COMMENT '用户ID',
    subscription_id INT UNSIGNED COMMENT '订阅ID',
    package_id INT UNSIGNED COMMENT '套餐ID',
    amount INT NOT NULL COMMENT '分发数量',
    distributed_at TIMESTAMP NOT NULL COMMENT '分发时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
);
```

### 8. 项目表 (projects)

```sql
CREATE TABLE projects (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    username VARCHAR(50) NOT NULL COMMENT '用户名',
    title VARCHAR(255) NOT NULL COMMENT '项目标题',
    description TEXT COMMENT '项目描述',
    genre VARCHAR(50) COMMENT '作品风格/类型',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_edited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后编辑时间',
    INDEX idx_user_id (user_id),
    INDEX idx_title (title),
    INDEX idx_username (username)
);
```

### 9. 大纲内容表 (outlines)

```sql
CREATE TABLE outlines (
    id INT PRIMARY KEY AUTO_INCREMENT,
    project_id INT NOT NULL COMMENT '项目ID',
    content TEXT NOT NULL COMMENT '大纲内容',
    current_version INT NOT NULL DEFAULT 1 COMMENT '当前版本号',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_project_id (project_id)
);
```

### 10. 版本历史表 (versions)

```sql
CREATE TABLE versions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    outline_id INT NOT NULL COMMENT '大纲ID',
    version_number INT NOT NULL COMMENT '版本号',
    content TEXT NOT NULL COMMENT '内容',
    is_ai_generated BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否AI生成',
    ai_style VARCHAR(50) COMMENT 'AI续写风格',
    word_limit INT COMMENT 'AI续写字数限制',
    tokens_used INT COMMENT '使用的token数量',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_outline_id (outline_id),
    UNIQUE KEY unique_outline_version (outline_id, version_number)
);
```

### 11. 文件表 (files)

```sql
CREATE TABLE files (
    id INT PRIMARY KEY AUTO_INCREMENT,
    filename VARCHAR(255) NOT NULL COMMENT '文件名',
    description TEXT COMMENT '文件描述',
    uploader VARCHAR(50) COMMENT '上传者用户名',
    uploader_id INT NOT NULL COMMENT '上传者ID',
    link VARCHAR(255) UNIQUE NOT NULL COMMENT '文件链接',
    upload_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
    download_counter INT NOT NULL DEFAULT 0 COMMENT '下载次数',
    INDEX idx_filename (filename),
    INDEX idx_uploader_id (uploader_id),
    INDEX idx_link (link),
    INDEX idx_uploader (uploader)
);
```

### 12. 系统选项表 (options)

```sql
CREATE TABLE options (
    `key` VARCHAR(255) PRIMARY KEY COMMENT '选项键',
    `value` TEXT COMMENT '选项值'
);
```

## 主要关系说明

1. 一个用户(users)可以有一个推荐码(referrals)
2. 一个用户(users)可以被另一个用户推荐(referral_uses)
3. 一个用户(users)可以订阅多个套餐(packages)，通过subscriptions表关联
4. 一个用户(users)可以有多个token记录(token_records)
5. 一个用户(users)可以获得多次token分发(token_distributions)
6. 一个用户(users)可以创建多个项目(projects)
7. 一个项目(projects)有一个大纲(outlines)
8. 一个大纲(outlines)有多个版本历史(versions)
9. 系统设置存储在options表中

## 索引设计考虑

1. 用户表对username、email、token、github_id和wechat_id添加索引，提高查询效率
2. 推荐码表对code添加索引，方便查询
3. 项目表对user_id和username添加索引，方便查询用户的所有项目
4. 大纲表和版本历史表对相关ID添加索引，优化关联查询
5. 文件表对uploader和uploader_id添加索引，方便查询用户上传的文件
6. 所有关联表都添加了相应的索引，提高关联查询效率

## 数据类型说明

1. 模型中的id字段在Go代码中通常定义为int或uint类型，在数据库中应选择INT或BIGINT类型
2. 模型中的布尔类型字段在数据库中应使用BOOLEAN或TINYINT(1)类型
3. 在实现中注意处理数据类型映射，特别是Go中的uint和数据库中的INT UNSIGNED或BIGINT UNSIGNED的映射关系

## 重要功能实现说明

1. Token余额管理：通过token_records表记录用户Token的变动，通过User表的token字段存储当前余额
2. 推荐码系统：通过referrals和referral_uses表实现推荐码功能，包括生成推荐码、使用推荐码和记录奖励
3. 版本控制：通过outlines和versions表实现大纲内容的版本管理，记录每次编辑和AI续写的历史
4. 会员订阅：通过packages和subscriptions表实现会员套餐订阅功能

## 数据维护建议

1. 定期备份数据库，特别是用户内容相关表
2. 对于长期不活跃的项目，可考虑归档处理
3. 对于token_records和token_distributions表，可能需要定期归档历史数据，保持表性能
4. 根据实际使用情况考虑对表进行分区，特别是随时间增长较快的表
