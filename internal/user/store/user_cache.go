package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"qahub/internal/user/model"

	"github.com/redis/go-redis/v9"
)

// userCacheStore 是一个为 UserStore 实现的装饰器，它使用 Redis 增加了缓存层。
type userCacheStore struct {
	redisClient *redis.Client // Redis 客户端
	next        UserStore     // 链中的下一个 store (例如，数据库 store)
	expiration  time.Duration // 缓存过期时间
}

// NewUserCacheStore 创建一个带有缓存装饰的 UserStore 新实例。
func NewUserCacheStore(redisClient *redis.Client, next UserStore) UserStore {
	return &userCacheStore{
		redisClient: redisClient,
		next:        next,
		expiration:  time.Hour, // 默认缓存过期时间设置为1小时
	}
}

// --- 缓存键生成函数 ---

// userKey 根据用户ID生成缓存键
func userKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

// usernameKey 根据用户名生成缓存键
func usernameKey(username string) string {
	return fmt.Sprintf("user:username:%s", username)
}

// --- 写操作与缓存失效方法 ---

// CreateUser 直接调用下一层的 CreateUser。缓存将在用户首次被读取时填充。
func (s *userCacheStore) CreateUser(user *model.User) (int64, error) {
	return s.next.CreateUser(user)
}

// UpdateUser 首先更新数据库，如果成功，则使缓存失效。
func (s *userCacheStore) UpdateUser(user *model.User) error {
	// 1. 首先更新数据库
	err := s.next.UpdateUser(user)
	if err != nil {
		return err
	}

	// 2. 如果数据库更新成功，则删除对应的缓存以保证数据一致性
	ctx := context.Background()
	s.redisClient.Del(ctx, userKey(user.ID))
	s.redisClient.Del(ctx, usernameKey(user.Username))

	return nil
}

// DeleteUser 首先从数据库删除，如果成功，则使缓存失效。
func (s *userCacheStore) DeleteUser(id int64) error {
	// 为了让 username 相关的缓存也失效，我们需要先获取用户信息
	user, err := s.next.GetUserByID(id)
	if err != nil {
		// 如果用户不存在，也可以直接返回，取决于业务需求
		return err
	}

	// 1. 首先从数据库删除
	err = s.next.DeleteUser(id)
	if err != nil {
		return err
	}

	// 2. 如果数据库删除成功，则删除对应的缓存
	ctx := context.Background()
	s.redisClient.Del(ctx, userKey(id))
	if user != nil {
		s.redisClient.Del(ctx, usernameKey(user.Username))
	}

	return nil
}

// --- 读穿透缓存方法 ---

// GetUserByID 实现了“读穿透”缓存逻辑。
func (s *userCacheStore) GetUserByID(id int64) (*model.User, error) {
	ctx := context.Background()
	key := userKey(id)

	// 1. 首先尝试从 Redis 缓存中获取
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == nil {
		// 缓存命中
		var user model.User
		if json.Unmarshal([]byte(val), &user) == nil {
			// log.Printf("Cache hit for id: %d", id)

			// 反序列化成功，直接返回结果
			return &user, nil
		}
	}

	// 2. 缓存未命中，从下一层 (数据库) 获取
	user, err := s.next.GetUserByID(id)
	if err != nil {
		return nil, err // 数据库查询出错
	}

	// 3. 将从数据库获取到的数据写入缓存，以便下次使用
	jsonData, _ := json.Marshal(user)
	s.redisClient.Set(ctx, key, jsonData, s.expiration)

	return user, nil
}

// GetUserByUsername 同样实现了“读穿透”缓存逻辑。
func (s *userCacheStore) GetUserByUsername(username string) (*model.User, error) {
	ctx := context.Background()
	key := usernameKey(username)

	// 1. 尝试从缓存获取
	val, err := s.redisClient.Get(ctx, key).Result()
	if err == nil {
		// 缓存命中
		var user model.User
		if json.Unmarshal([]byte(val), &user) == nil {
			// log.Printf("Cache hit for username: %s", username)
			return &user, nil
		}
	}

	// 2. 缓存未命中，从数据库获取
	user, err := s.next.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	// 3. 写回缓存
	jsonData, _ := json.Marshal(user)
	s.redisClient.Set(ctx, key, jsonData, s.expiration)
	// 为了数据一致性，最好也根据ID再缓存一份
	s.redisClient.Set(ctx, userKey(user.ID), jsonData, s.expiration)

	return user, nil
}

// GetUserByEmail 在此示例中未被缓存，它会直接穿透到下一层。
func (s *userCacheStore) GetUserByEmail(email string) (*model.User, error) {
	return s.next.GetUserByEmail(email)
}
