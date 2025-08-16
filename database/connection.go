package database

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"alkaukaba-backend/config"

	driver "gorm.io/driver/mysql"        // 👉 alias 'driver' untuk GORM
	"gorm.io/gorm"

	gossh "golang.org/x/crypto/ssh"
	mysql "github.com/go-sql-driver/mysql" // 👉 alias 'mysql' untuk go-sql-driver
)

var DB *gorm.DB

func ConnectDB(cfg config.Config) {
	// === Setup SSH client ===
	key, err := ioutil.ReadFile(cfg.SSHKey)
	if err != nil {
		log.Fatalf("Failed to read SSH key: %v", err)
	}
	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to parse SSH key: %v", err)
	}

	sshConfig := &gossh.ClientConfig{
		User: cfg.SSHUser,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	sshAddress := fmt.Sprintf("%s:%s", cfg.SSHHost, cfg.SSHPort)
	sshConn, err := gossh.Dial("tcp", sshAddress, sshConfig)
	if err != nil {
		log.Fatalf("Failed to connect SSH: %v", err)
	}

	// === Register custom dialer untuk MySQL ===
	mysql.RegisterDial("mysql+tcp", func(addr string) (net.Conn, error) {
		return sshConn.Dial("tcp", addr)
	})

	// === DSN via SSH ===
	dbAddress := fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort)
	dsn := fmt.Sprintf("%s:%s@mysql+tcp(%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		cfg.DBUser, cfg.DBPass, dbAddress, cfg.DBName)

	DB, err = gorm.Open(driver.Open(dsn), &gorm.Config{}) // 👉 pakai 'driver'
	if err != nil {
		log.Fatalf("Failed to connect database via SSH: %v", err)
	}

	log.Println("✅ Connected to MySQL via SSH Tunnel")
}
