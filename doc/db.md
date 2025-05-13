# 网文大纲续写工具数据库设计

## 数据库选择
推荐使用MySQL或PostgreSQL关系型数据库，以确保数据完整性和支持复杂查询。

## 表结构设计

### 1. 用户表 (users)

```sql
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(255),
    referral_code VARCHAR(20) NOT NULL UNIQUE,
    token_balance INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1:活跃, 0:禁用',
    INDEX idx_referral_code (referral_code)
);
```

### 2. 套餐表 (plans)

```sql
CREATE TABLE plans (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    duration_days INT NOT NULL COMMENT '套餐有效期(天)',
    monthly_tokens INT NOT NULL COMMENT '每月赠送token数量',
    is_permanent BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否永久套餐',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1:可用, 0:下架'
);
```

### 3. 用户套餐表 (user_plans)

```sql
CREATE TABLE user_plans (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    plan_id INT NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (plan_id) REFERENCES plans(id),
    INDEX idx_user_id (user_id),
    INDEX idx_plan_id (plan_id)
);
```

### 4. 项目表 (projects)

```sql
CREATE TABLE projects (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    genre VARCHAR(50) COMMENT '作品风格/类型',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_edited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);
```

### 5. 大纲内容表 (outlines)

```sql
CREATE TABLE outlines (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    project_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    current_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    INDEX idx_project_id (project_id)
);
```

### 6. 版本历史表 (versions)

```sql
CREATE TABLE versions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    outline_id BIGINT NOT NULL,
    version_number INT NOT NULL,
    content TEXT NOT NULL,
    is_ai_generated BOOLEAN NOT NULL DEFAULT FALSE,
    ai_style VARCHAR(50) COMMENT 'AI续写风格',
    word_limit INT COMMENT 'AI续写字数限制',
    tokens_used INT COMMENT '使用的token数',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (outline_id) REFERENCES outlines(id),
    INDEX idx_outline_id (outline_id),
    UNIQUE KEY unique_outline_version (outline_id, version_number)
);
```

### 7. Token记录表 (token_records)

```sql
CREATE TABLE token_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    amount INT NOT NULL COMMENT '正值为增加,负值为消费',
    balance INT NOT NULL COMMENT '变动后余额',
    record_type TINYINT NOT NULL COMMENT '1:套餐赠送, 2:推荐奖励, 3:续写消费, 4:充值',
    related_id BIGINT COMMENT '相关记录ID',
    description VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);
```

### 8. 推荐记录表 (referrals)

```sql
CREATE TABLE referrals (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    referrer_id BIGINT NOT NULL COMMENT '推荐人ID',
    referred_id BIGINT NOT NULL COMMENT '被推荐人ID',
    reward_tokens INT NOT NULL DEFAULT 0 COMMENT '获得的奖励token',
    is_rewarded BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (referrer_id) REFERENCES users(id),
    FOREIGN KEY (referred_id) REFERENCES users(id),
    INDEX idx_referrer_id (referrer_id),
    INDEX idx_referred_id (referred_id),
    UNIQUE KEY unique_referral (referred_id)
);
```

### 9. 导出记录表 (exports)

```sql
CREATE TABLE exports (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    project_id BIGINT NOT NULL,
    file_format VARCHAR(20) NOT NULL COMMENT 'TXT, DOCX, PDF等',
    file_url VARCHAR(255),
    file_size INT COMMENT '文件大小(KB)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (project_id) REFERENCES projects(id),
    INDEX idx_user_id (user_id),
    INDEX idx_project_id (project_id)
);
```

## 主要关系说明

1. 一个用户(users)可以有多个项目(projects)
2. 一个项目(projects)有一个大纲(outlines)
3. 一个大纲(outlines)有多个版本历史(versions)
4. 一个用户(users)可以购买多个套餐(plans)，通过user_plans关联
5. 一个用户(users)可以推荐多个新用户，通过referrals关联
6. 用户的token变动记录在token_records表中

## 索引设计考虑

1. 用户表使用用户名、邮箱作为唯一索引，推荐码添加普通索引方便查询
2. 项目表和大纲表按用户ID和项目ID建立索引，方便按用户查询项目
3. 版本历史表按大纲ID建立索引，同时维护版本号唯一性
4. token记录表和推荐记录表按用户ID建立索引，方便查询用户相关记录

## 数据维护建议

1. 定期备份数据库，特别是用户内容相关表
2. 对于长期不活跃的项目，可考虑归档处理
3. 对于token_records表，可能需要定期归档历史数据，保持表性能
