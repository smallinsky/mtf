package components

import (
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

type MySQL struct {
	Pass     string
	Port     string
	DB       []string
	Hostname string
	Network  string
}

func (c *MySQL) Start() {
	log.Println("--- INIT MYSQL START ----")
	cmd := `docker run -d --rm --network=mtf_net --name mysql_mtf --env MYSQL_ROOT_PASSWORD=test --env MYSQL_DATABASE=test_db -p 3306:3306 mysql --default-authentication-plugin=mysql_native_password`
	run(cmd)
	log.Println("--- INIT MYSQL DONE ----")
}

func (c *MySQL) Stop() {
	log.Println("--- DEL MYSQL START ---")
	run("docker stop mysql_mtf")
	log.Println("--- DEL MYSQL DONE ---")
}

func (c *MySQL) Ready() {
	waitForOpenPort("mysql_mtf", "3306")
}

func run(s string) error {
	args := strings.Split(s, " ")
	buff, err := exec.Command(args[0], args[1:len(args)]...).Output()
	if err != nil {
		log.Printf("[ERROR] cmd run: %v , \n", err)
		return err
	}
	log.Printf("[DEBUG] output: %v\n", string(buff))
	return nil
}

func waitForOpenPort(host, port string) {
	for {
		time.Sleep(time.Millisecond * 50)
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Millisecond*500)
		if err != nil {
			continue
		}
		if conn != nil {
			conn.Close()
			break
		}
	}
}
