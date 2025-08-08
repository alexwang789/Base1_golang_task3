package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	 "github.com/jmoiron/sqlx"
)

// Employee 结构体映射 employees 表
type Employee struct {
	ID         int    `db:"id"`
	Name       string `db:"name"`
	Department string `db:"department"`
	Salary     int    `db:"salary"`
}

func main() {
	// 初始化数据库连接
	db, err := initDB()
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 1. 查询技术部所有员工
	fmt.Println("技术部员工列表:")
	techEmployees, err := getEmployeesByDepartment(db, "技术部")
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		for _, emp := range techEmployees {
			fmt.Printf("- ID: %d, 姓名: %s, 部门: %s, 薪资: %d\n", 
				emp.ID, emp.Name, emp.Department, emp.Salary)
		}
	}

	// 2. 查询工资最高的员工
	fmt.Println("\n工资最高的员工:")
	topEarner, err := getHighestPaidEmployee(db)
	if err != nil {
		log.Printf("查询失败: %v", err)
	} else {
		fmt.Printf("- ID: %d, 姓名: %s, 部门: %s, 薪资: %d\n", 
			topEarner.ID, topEarner.Name, topEarner.Department, topEarner.Salary)
	}
}

// 初始化数据库连接
func initDB() (*sqlx.DB, error) {
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
		dbName = "company_db"
	}
	
	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", 
		dbUser, dbPass, dbHost, dbPort, dbName)
	
	// 创建数据库连接
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}
	
	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	fmt.Println("✅ 数据库连接成功")
	return db, nil
}

// 1. 查询指定部门的所有员工
func getEmployeesByDepartment(db *sqlx.DB, department string) ([]Employee, error) {
	query := `
		SELECT id, name, department, salary
		FROM employees
		WHERE department = ?
	`
	
	var employees []Employee
	err := db.Select(&employees, query, department)
	if err != nil {
		return nil, fmt.Errorf("查询部门员工失败: %w", err)
	}
	
	if len(employees) == 0 {
		return nil, fmt.Errorf("部门 '%s' 没有员工", department)
	}
	
	return employees, nil
}

// 2. 查询工资最高的员工
func getHighestPaidEmployee(db *sqlx.DB) (Employee, error) {
	query := `
		SELECT id, name, department, salary
		FROM employees
		ORDER BY salary DESC
		LIMIT 1
	`
	
	var employee Employee
	err := db.Get(&employee, query)
	if err != nil {
		return Employee{}, fmt.Errorf("查询最高薪资员工失败: %w", err)
	}
	
	return employee, nil
}

// 可选：获取所有最高薪资员工（处理并列情况）
func getAllHighestPaidEmployees(db *sqlx.DB) ([]Employee, error) {
	query := `
		SELECT id, name, department, salary
		FROM employees
		WHERE salary = (SELECT MAX(salary) FROM employees)
	`
	
	var employees []Employee
	err := db.Select(&employees, query)
	if err != nil {
		return nil, fmt.Errorf("查询所有最高薪资员工失败: %w", err)
	}
	
	return employees, nil
}
