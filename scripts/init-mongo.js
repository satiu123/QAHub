// MongoDB 初始化脚本 - 创建 notification_db 数据库用户
db = db.getSiblingDB('notification_db');

// 创建用户并授予 notification_db 数据库的读写权限
db.createUser({
    user: 'qahub_admin',
    pwd: '12345678',
    roles: [
        {
            role: 'readWrite',
            db: 'notification_db'
        }
    ]
});

print('MongoDB initialization script completed - user qahub_admin created for notification_db');