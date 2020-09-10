package docker

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"log"
	"sync"
)

func Postgres(ctx context.Context, wg *sync.WaitGroup) (ip string, err error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return ip, err
	}
	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11.2",
		Env:        []string{"POSTGRES_DB=camunda", "POSTGRES_PASSWORD=camunda", "POSTGRES_USER=camunda"},
	}, func(config *docker.HostConfig) {
		config.Tmpfs = map[string]string{"/var/lib/postgresql/data": "rw"}
	})
	if err != nil {
		return "", err
	}
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()
	ip = container.Container.NetworkSettings.IPAddress
	conStr := fmt.Sprintf("postgres://camunda:camunda@%s:5432/%s?sslmode=disable", ip, "camunda")
	err = pool.Retry(func() error {
		var err error
		log.Println("try connecting to pg")
		db, err := sql.Open("postgres", conStr)
		if err != nil {
			log.Println(err)
			return err
		}
		err = db.Ping()
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})
	return
}
