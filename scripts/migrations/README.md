# 数据库迁移说明

## 统一迁移目录

为了解决不同服务间的依赖关系问题，所有数据库迁移文件已经整合到 `scripts/migrations/all/` 目录下。

## 迁移文件顺序

迁移文件按照依赖关系进行排序：

1. `000001_create_users_table` - 创建用户表（基础表，无依赖）
2. `000002_create_questions_table` - 创建问题表（依赖用户表）
3. `000003_create_answers_table` - 创建答案表（依赖问题表和用户表）

## 使用方法

### 执行所有迁移
```bash
migrate -database "mysql://root:12345678@tcp(localhost:3307)/qahub?charset=utf8mb4&parseTime=True&loc=Local" -path=scripts/migrations/all up
```

### 回滚迁移
```bash
# 回滚一步
migrate -database "mysql://root:12345678@tcp(localhost:3307)/qahub?charset=utf8mb4&parseTime=True&loc=Local" -path=scripts/migrations/all down 1

# 回滚到指定版本
migrate -database "mysql://root:12345678@tcp(localhost:3307)/qahub?charset=utf8mb4&parseTime=True&loc=Local" -path=scripts/migrations/all goto 1
```

### 检查迁移状态
```bash
migrate -database "mysql://root:12345678@tcp(localhost:3307)/qahub?charset=utf8mb4&parseTime=True&loc=Local" -path=scripts/migrations/all version
```

## 注意事项

1. **不要再使用** `scripts/migrations/user/` 和 `scripts/migrations/qa/` 目录中的旧迁移文件
2. 所有新的迁移都应该添加到 `scripts/migrations/all/` 目录下
3. 新迁移的编号应该从 `000004` 开始
4. 确保新迁移考虑到表之间的依赖关系

## 外键约束关系

- `questions.user_id` → `users.id`
- `answers.user_id` → `users.id`
- `answers.question_id` → `questions.id`