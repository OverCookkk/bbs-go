package cache

import (
	"errors"
	"time"

	"bbs-go/model"
	"bbs-go/repositories"

	"github.com/goburrow/cache"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
)

type userCache struct {
	cache            cache.LoadingCache
	scoreRankCache   cache.LoadingCache
	checkInRankCache cache.LoadingCache
}

var UserCache = newUserCache()

func newUserCache() *userCache {
	return &userCache{
		cache: cache.NewLoadingCache(
			func(key cache.Key) (value cache.Value, e error) {
				value = repositories.UserRepository.Get(sqls.DB(), key2Int64(key))
				if value == nil {
					e = errors.New("数据不存在")
				}
				return
			},
			cache.WithMaximumSize(1000),
			cache.WithExpireAfterAccess(30*time.Minute),
		),
		scoreRankCache: cache.NewLoadingCache(
			func(key cache.Key) (value cache.Value, e error) {
				value = repositories.UserRepository.Find(sqls.DB(), sqls.NewCnd().Desc("score").Limit(10))
				if value == nil {
					e = errors.New("数据不存在")
				}
				return
			},
			cache.WithMaximumSize(10),
			cache.WithRefreshAfterWrite(10*time.Minute),
		),
		checkInRankCache: cache.NewLoadingCache(
			func(key cache.Key) (value cache.Value, e error) {
				today := dates.GetDay(time.Now())
				value = repositories.CheckInRepository.Find(sqls.DB(),
					sqls.NewCnd().Eq("latest_day_name", today).Asc("update_time").Limit(10))
				return
			},
			cache.WithMaximumSize(10),
			cache.WithExpireAfterAccess(1*time.Hour),
		),
	}
}

// Get 调用c.cache.Get(userId)，首先检查指定键的值是否已经存在于缓存中，如果存在且未过期，则直接返回缓存中的值，而无需执行回调函数。
// 只有当缓存中不存在或已过期时，就会触发执行userCache对象里cache里面的回调函数，也就是会去数据库中查询用户信息。
func (c *userCache) Get(userId int64) *model.User {
	if userId <= 0 {
		return nil
	}
	val, err := c.cache.Get(userId)
	if err != nil {
		return nil
	}
	return val.(*model.User)
}

func (c *userCache) Invalidate(userId int64) {
	c.cache.Invalidate(userId)
}

func (c *userCache) GetScoreRank() []model.User {
	val, err := c.scoreRankCache.Get("data")
	if err != nil {
		return nil
	}
	return val.([]model.User)
}

func (c *userCache) GetCheckInRank() []model.CheckIn {
	today := dates.GetDay(time.Now())
	val, err := c.checkInRankCache.Get(today)
	if err != nil {
		return nil
	}
	return val.([]model.CheckIn)
}

func (c *userCache) RefreshCheckInRank() {
	c.checkInRankCache.Refresh(dates.GetDay(time.Now()))
}
