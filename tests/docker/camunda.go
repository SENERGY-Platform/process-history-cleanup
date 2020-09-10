package docker

import (
	"context"
	"errors"
	"github.com/ory/dockertest/v3"
	"log"
	"net/http"
	"sync"
)

func Camunda(ctx context.Context, wg *sync.WaitGroup, dbIp string) (url string, err error) {
	log.Println("start camunda")
	pool, err := dockertest.NewPool("")
	if err != nil {
		return url, err
	}
	container, err := pool.Run("fgseitsrancher.wifa.intern.uni-leipzig.de:5000/process-engine", "prod", []string{
		"DB_USERNAME=camunda",
		"DB_URL=jdbc:postgresql://" + dbIp + ":5432/camunda",
		"DB_PORT=5432",
		"DB_PASSWORD=camunda",
		"DB_NAME=camunda",
		"DB_HOST=" + dbIp,
		"DB_DRIVER=org.postgresql.Driver",
	})
	if err != nil {
		return "", err
	}
	go Dockerlog(pool, ctx, container, "CAMUNDA")
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("DEBUG: remove container " + container.Container.Name)
		container.Close()
		wg.Done()
	}()

	url = "http://" + container.Container.NetworkSettings.IPAddress + ":8080"

	err = pool.Retry(func() error {
		log.Println("DEBUG: try to connection to camunda")
		resp, err := http.Get(url + "/engine-rest/case-instance/count")
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			return errors.New(resp.Status)
		}
		return nil
	})
	return url, err
}
