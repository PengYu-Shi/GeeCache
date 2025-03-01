package DB

import (
	"GeeCacheNode/conf"
	"GeeCacheNode/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

var DB *gorm.DB

func InitDB(config conf.MysqlConfig) (*gorm.DB, error) {
	host := config.DbHost
	port := config.DbPort
	database := config.DbName
	username := config.DbUser
	password := config.DbPass
	charset := config.Charset
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		username, password, host, port, database, charset)
	db, err := gorm.Open(mysql.Open(args))
	if err != nil {
		panic("failed to connect database, err: " + err.Error())
	}
	err = db.AutoMigrate(&model.Score{})
	if err != nil {
		return nil, err
	}
	DB = db
	return db, nil
}

func GetDB() *gorm.DB {
	return DB
}

//	func (db *DB) GetByKey(key string) ([]byte, error) {
//		var score model.Score
//		result := db.Db.Where("key = ?", key).First(&score)
//		if result.Error != nil {
//			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
//				return nil, nil // 或者返回一个自定义的错误
//			}
//			return nil, result.Error
//		}
//		return []byte(score.Score), nil
//	}

const batchSize = 1000

func InsertData(db *gorm.DB) error {
	rand.Seed(time.Now().UnixNano())

	names := generateUniqueNumbers(100000)
	totalRecords := len(names)

	for i := 0; i < totalRecords; i += batchSize {
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		scores := make([]model.Score, end-i)
		for j, name := range names[i:end] {
			score := strconv.Itoa(rand.Intn(100)) // 生成0到99之间的随机数作为score
			scores[j] = model.Score{
				Key:   name,
				Score: score,
			}
		}

		// 批量插入数据
		result := db.Create(&scores)
		if result.Error != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %w", i, end-1, result.Error)
		}

		fmt.Printf("Successfully inserted batch %d-%d\n", i, end-1)
	}

	return nil
}

// generateUniqueNumbers 生成指定数量的唯一数字
func generateUniqueNumbers(count int) []string {
	numbers := make(map[string]struct{})
	for len(numbers) < count {
		number := strconv.Itoa(rand.Intn(100000) + 1) // 生成1到100000之间的唯一数字
		numbers[number] = struct{}{}
	}

	uniqueNumbers := make([]string, 0, count)
	for number := range numbers {
		uniqueNumbers = append(uniqueNumbers, number)
	}

	return uniqueNumbers
}
