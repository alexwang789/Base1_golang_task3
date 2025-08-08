package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 1. 模型定义

// User 用户模型
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Name         string    `gorm:"size:100;not null;uniqueIndex"`
	Email        string    `gorm:"size:100;not null;uniqueIndex"`
	Password     string    `gorm:"size:255;not null"`
	ArticleCount int       `gorm:"default:0"` // 文章数量统计
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Posts        []Post // 一对多关系: 用户 -> 文章
}

// Post 文章模型
type Post struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text;not null"`
	CommentStatus string    `gorm:"size:20;default:'无评论'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UserID        uint     // 外键
	User          User     `gorm:"foreignKey:UserID"` // 多对一关系: 文章 -> 用户
	Comments      []Comment // 一对多关系: 文章 -> 评论
}

// Comment 评论模型
type Comment struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	PostID    uint // 外键
	Post      Post `gorm:"foreignKey:PostID"` // 多对一关系: 评论 -> 文章
	UserID    uint // 外键
	User      User `gorm:"foreignKey:UserID"` // 多对一关系: 评论 -> 用户
}

func main() {
	// 初始化数据库连接
	db, err := initDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer closeDB(db)

	// 自动迁移创建表
	if err := db.AutoMigrate(&User{}, &Post{}, &Comment{}); err != nil {
		log.Fatalf("表创建失败: %v", err)
	}
	fmt.Println("✅ 数据表已创建")

	// 创建测试数据
	if err := createTestData(db); err != nil {
		log.Fatalf("创建测试数据失败: %v", err)
	}

	// 2. 关联查询
	// 查询用户1的所有文章及其评论
	fmt.Println("\n查询用户1的所有文章及其评论:")
	if err := queryUserPostsWithComments(db, 1); err != nil {
		log.Printf("查询失败: %v", err)
	}

	// 查询评论数量最多的文章
	fmt.Println("\n查询评论数量最多的文章:")
	if err := queryMostCommentedPost(db); err != nil {
		log.Printf("查询失败: %v", err)
	}

	// 3. 钩子函数测试
	// 创建新文章测试钩子
	fmt.Println("\n创建新文章测试钩子:")
	newPost := Post{
		Title:   "钩子函数测试文章",
		Content: "测试创建文章时自动更新用户文章数量",
		UserID:  1,
	}
	if err := db.Create(&newPost).Error; err != nil {
		log.Printf("创建文章失败: %v", err)
	} else {
		fmt.Println("✅ 文章创建成功")
	}

	// 删除评论测试钩子
	fmt.Println("\n删除评论测试钩子:")
	var comment Comment
	if err := db.First(&comment).Error; err != nil {
		log.Printf("获取评论失败: %v", err)
	} else {
		if err := db.Delete(&comment).Error; err != nil {
			log.Printf("删除评论失败: %v", err)
		} else {
			fmt.Println("✅ 评论删除成功")
		}
	}

	// 显示最终用户和文章状态
	fmt.Println("\n最终用户和文章状态:")
	if err := showFinalStatus(db); err != nil {
		log.Printf("查询失败: %v", err)
	}
}

// 初始化数据库连接
func initDB() (*gorm.DB, error) {
	// 从环境变量获取数据库配置
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	
	// 设置默认值
	if dbUser == "" || dbPass == "" {
		dbUser = "root"
		dbPass = "password"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbName == "" {
		dbName = "blog_db"
	}
	
	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", 
		dbUser, dbPass, dbHost, dbPort, dbName)
	
	// 配置GORM日志
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	// 创建数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	
	// 获取通用数据库对象 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}
	
	// 配置连接池
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	fmt.Println("🚀 数据库连接成功")
	return db, nil
}

// 关闭数据库连接
func closeDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("获取数据库连接失败: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %v", err)
	}
}

// 创建测试数据
func createTestData(db *gorm.DB) error {
	// 创建用户
	users := []User{
		{Name: "张三", Email: "zhangsan@example.com", Password: "pass123"},
		{Name: "李四", Email: "lisi@example.com", Password: "pass456"},
	}
	
	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}
	}
	
	// 创建文章
	posts := []Post{
		{Title: "Go语言入门", Content: "Go语言基础教程...", UserID: users[0].ID},
		{Title: "GORM使用指南", Content: "GORM高级技巧...", UserID: users[0].ID},
		{Title: "Web开发实践", Content: "使用Go开发Web应用...", UserID: users[1].ID},
	}
	
	for i := range posts {
		if err := db.Create(&posts[i]).Error; err != nil {
			return fmt.Errorf("创建文章失败: %w", err)
		}
	}
	
	// 创建评论
	comments := []Comment{
		{Content: "好文章！", PostID: posts[0].ID, UserID: users[1].ID},
		{Content: "学到了很多", PostID: posts[0].ID, UserID: users[0].ID},
		{Content: "期待更多内容", PostID: posts[1].ID, UserID: users[1].ID},
	}
	
	for i := range comments {
		if err := db.Create(&comments[i]).Error; err != nil {
			return fmt.Errorf("创建评论失败: %w", err)
		}
	}
	
	fmt.Println("✅ 测试数据创建成功")
	return nil
}

// 2.1 查询用户的所有文章及其评论
func queryUserPostsWithComments(db *gorm.DB, userID uint) error {
	var user User
	
	// 预加载文章和文章的评论
	err := db.Preload("Posts.Comments").First(&user, userID).Error
	if err != nil {
		return fmt.Errorf("查询用户失败: %w", err)
	}
	
	fmt.Printf("用户 %s 的文章:\n", user.Name)
	for i, post := range user.Posts {
		fmt.Printf("  %d. %s (评论数: %d)\n", i+1, post.Title, len(post.Comments))
		for j, comment := range post.Comments {
			fmt.Printf("    - %d. %s\n", j+1, comment.Content)
		}
	}
	
	return nil
}

// 2.2 查询评论数量最多的文章
func queryMostCommentedPost(db *gorm.DB) error {
	var post Post
	
	// 使用子查询获取评论最多的文章
	err := db.Raw(`
		SELECT posts.*
		FROM posts
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comment_count
			FROM comments
			GROUP BY post_id
		) AS comment_counts ON posts.id = comment_counts.post_id
		ORDER BY comment_counts.comment_count DESC
		LIMIT 1
	`).Scan(&post).Error
	
	if err != nil {
		return fmt.Errorf("查询失败: %w", err)
	}
	
	// 获取评论数量
	var commentCount int64
	db.Model(&Comment{}).Where("post_id = ?", post.ID).Count(&commentCount)
	
	fmt.Printf("评论最多的文章: %s (ID: %d, 评论数: %d)\n", 
		post.Title, post.ID, commentCount)
	
	return nil
}

// 3.1 Post 钩子函数 - 创建文章后更新用户文章数量
func (p *Post) AfterCreate(tx *gorm.DB) error {
	// 更新用户的文章数量
	result := tx.Model(&User{}).Where("id = ?", p.UserID).
		Update("article_count", gorm.Expr("article_count + ?", 1))
	
	if result.Error != nil {
		return result.Error
	}
	
	fmt.Printf("✅ 用户 %d 的文章数量已更新\n", p.UserID)
	return nil
}

// 3.2 Comment 钩子函数 - 删除评论后检查文章评论状态
func (c *Comment) AfterDelete(tx *gorm.DB) error {
	// 获取文章当前的评论数量
	var commentCount int64
	if err := tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&commentCount).Error; err != nil {
		return err
	}
	
	// 更新文章评论状态
	newStatus := "有评论"
	if commentCount == 0 {
		newStatus = "无评论"
	}
	
	if err := tx.Model(&Post{}).Where("id = ?", c.PostID).
		Update("comment_status", newStatus).Error; err != nil {
		return err
	}
	
	fmt.Printf("✅ 文章 %d 的评论状态已更新为: %s\n", c.PostID, newStatus)
	return nil
}

// 显示最终状态
func showFinalStatus(db *gorm.DB) error {
	// 查询所有用户
	var users []User
	if err := db.Find(&users).Error; err != nil {
		return err
	}
	
	fmt.Println("用户文章数量统计:")
	for _, user := range users {
		fmt.Printf("- %s: %d 篇文章\n", user.Name, user.ArticleCount)
	}
	
	// 查询所有文章
	var posts []Post
	if err := db.Find(&posts).Error; err != nil {
		return err
	}
	
	fmt.Println("\n文章评论状态:")
	for _, post := range posts {
		fmt.Printf("- %s: %s\n", post.Title, post.CommentStatus)
	}
	
	return nil
}
