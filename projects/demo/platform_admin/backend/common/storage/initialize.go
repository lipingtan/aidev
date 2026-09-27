/*
 * @Author: zhangwenjian
 * @Date: 2025/04/13 22:03
 * @Last Modified by: zhangwenjian
 * @Last Modified time: 2025/04/13 22:03
 */

package storage

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/go-admin-team/go-admin-core/sdk"
	"github.com/go-admin-team/go-admin-core/sdk/config"
	"github.com/go-admin-team/go-admin-core/sdk/pkg/captcha"
)

// SettingsFile 配置文件路径
var SettingsFile = "config/settings.yml"

// CacheConfig 缓存配置（从 settings.yml 解析）
type CacheConfig struct {
	Driver   string `yaml:"driver"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// loadCacheConfig 从 settings.yml 加载 cache 配置
func loadCacheConfig() *CacheConfig {
	data, err := os.ReadFile(SettingsFile)
	if err != nil {
		log.Printf("warning: failed to read settings file %s: %v", SettingsFile, err)
		return nil
	}

	var settings struct {
		Settings struct {
			Cache *CacheConfig `yaml:"cache"`
		} `yaml:"settings"`
	}

	if err := yaml.Unmarshal(data, &settings); err != nil {
		log.Printf("warning: failed to parse settings file: %v", err)
		return nil
	}

	return settings.Settings.Cache
}

// Setup 配置storage组件
func Setup() {
	setupCache()
	setupCaptcha()
	setupQueue()
}

func setupCache() {
	// 尝试从 settings.yml 加载 cache 配置
	cacheCfg := loadCacheConfig()

	if cacheCfg != nil && cacheCfg.Driver == "redis" {
		// 使用 Redis 适配器
		addr := cacheCfg.Addr
		if addr == "" {
			addr = "127.0.0.1:6379"
		}
		adapter, err := NewRedisAdapter(addr, cacheCfg.Password, cacheCfg.DB)
		if err != nil {
			log.Printf("warning: failed to create Redis adapter: %v, falling back to memory cache", err)
			adapter = nil
		}
		if adapter != nil {
			sdk.Runtime.SetCacheAdapter(adapter)
			log.Printf("cache: using Redis adapter at %s db=%d", addr, cacheCfg.DB)
			return
		}
	}

	// 默认使用内存缓存
	cacheAdapter, err := config.CacheConfig.Setup()
	if err != nil {
		log.Fatalf("cache setup error, %s\n", err.Error())
	}
	sdk.Runtime.SetCacheAdapter(cacheAdapter)
	log.Printf("cache: using memory adapter")
}

func setupCaptcha() {
	captcha.SetStore(captcha.NewCacheStore(sdk.Runtime.GetCacheAdapter(), 600))
}

func setupQueue() {
	if config.QueueConfig.Empty() {
		return
	}
	if q := sdk.Runtime.GetQueueAdapter(); q != nil {
		q.Shutdown()
	}
	queueAdapter, err := config.QueueConfig.Setup()
	if err != nil {
		log.Fatalf("queue setup error, %s\n", err.Error())
	}
	sdk.Runtime.SetQueueAdapter(queueAdapter)
	go queueAdapter.Run()
}
