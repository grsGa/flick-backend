package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const repositoryTemplate = `
// Repository 数据存储层
type Repository struct {
	*PostgresRepository
	logger zerolog.Logger
}

// NewRepository 创建一个新的Repository实例
func NewRepository(db *gorm.DB, logger zerolog.Logger) *Repository {
	// 创建PostgreSQL仓库实现
	postgresRepo := NewPostgresRepository(db)
	
	return &Repository{
		PostgresRepository: postgresRepo,
		logger: logger,
	}
}
`

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})
	logger := log.With().Str("script", "fix_repositories").Logger()

	// 服务列表
	services := []string{"user", "content", "notification", "interaction", "recommendation"}

	// 检查每个服务
	for _, service := range services {
		checkServiceRepository(service, logger)
	}
}

func checkServiceRepository(serviceName string, logger zerolog.Logger) {
	logger.Info().Msgf("检查服务 %s 的Repository实现", serviceName)

	// 仓库文件路径
	repoFilePath := filepath.Join("services", serviceName, "repository", "repository.go")
	postgresFilePath := filepath.Join("services", serviceName, "repository", "postgres.go")

	// 检查文件是否存在
	if _, err := os.Stat(repoFilePath); os.IsNotExist(err) {
		logger.Warn().Msgf("服务 %s 的Repository文件不存在: %s", serviceName, repoFilePath)
		return
	}

	if _, err := os.Stat(postgresFilePath); os.IsNotExist(err) {
		logger.Warn().Msgf("服务 %s 的PostgresRepository文件不存在: %s", serviceName, postgresFilePath)
		// 如果没有PostgresRepository，可能需要进一步处理
	}

	// 读取文件内容
	repoContent, err := ioutil.ReadFile(repoFilePath)
	if err != nil {
		logger.Error().Err(err).Msgf("无法读取 %s", repoFilePath)
		return
	}

	// 检查是否包含正确的Repository实现
	repoContentStr := string(repoContent)
	
	// 检查是否使用PostgresRepository
	if !strings.Contains(repoContentStr, "*PostgresRepository") {
		logger.Warn().Msgf("服务 %s 的Repository未嵌入PostgresRepository", serviceName)
		// 打印需要修改的内容
		fmt.Printf("\n=== 服务 %s 需要修改Repository实现 ===\n", serviceName)
		fmt.Printf("// 当前实现:\n")
		
		// 提取Repository结构体定义
		repoStructStart := strings.Index(repoContentStr, "type Repository struct")
		if repoStructStart == -1 {
			logger.Error().Msgf("无法找到Repository结构体定义")
			return
		}
		
		repoStructEnd := strings.Index(repoContentStr[repoStructStart:], "}")
		if repoStructEnd == -1 {
			logger.Error().Msgf("无法找到Repository结构体结束")
			return
		}
		
		repoStruct := repoContentStr[repoStructStart : repoStructStart+repoStructEnd+1]
		fmt.Println(repoStruct)
		
		// 提取NewRepository函数定义
		newRepoStart := strings.Index(repoContentStr, "func NewRepository")
		if newRepoStart == -1 {
			logger.Error().Msgf("无法找到NewRepository函数定义")
			return
		}
		
		newRepoEnd := strings.Index(repoContentStr[newRepoStart:], "}")
		if newRepoEnd == -1 {
			logger.Error().Msgf("无法找到NewRepository函数结束")
			return
		}
		
		newRepoFunc := repoContentStr[newRepoStart : newRepoStart+newRepoEnd+1]
		fmt.Println(newRepoFunc)
		
		// 打印推荐修改
		fmt.Printf("\n// 推荐修改为:\n")
		fmt.Println(repositoryTemplate)
	} else {
		logger.Info().Msgf("服务 %s 的Repository已经嵌入PostgresRepository", serviceName)
	}
} 