package crawler

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/griddis/atlant_test/tools/logging"
)

type Crawler struct {
	Status      bool
	Logger      *logging.Logger
	concurrency int
	Workers     []*Worker
	Input       chan string
	Output      chan string
}

func NewCrawler(ctx context.Context, concurrency int, client *http.Client) *Crawler {
	logger := logging.FromContext(ctx)
	craw := Crawler{
		Status:      true,
		Logger:      logger,
		concurrency: concurrency,
		Workers:     make([]*Worker, concurrency),
		Input:       make(chan string),
		Output:      make(chan string),
	}
	for i := 0; i < concurrency; i++ {
		craw.Workers[i] = NewWorker(logger.With("worker", i), client, &craw)
		go craw.Workers[i].Run(craw.Input, craw.Output)
	}
	return &craw
}

func (c *Crawler) Close() {
	for i := 0; i < c.concurrency; i++ {
		c.Workers[i].Close()
	}
	close(c.Input)
	close(c.Output)
}

func (c *Crawler) ChangeStatus(s bool) {
	c.Status = s
}

type Worker struct {
	logger  *logging.Logger
	client  *http.Client
	cancel  context.CancelFunc
	crawler *Crawler
	status  bool
}

func NewWorker(logger *logging.Logger, client *http.Client, craw *Crawler) *Worker {
	return &Worker{
		logger:  logger,
		status:  true,
		crawler: craw,
		client:  client,
	}
}

func (w *Worker) Run(input chan string, output chan string) {
	w.logger.Debug("status", "start")

	for url := range input {
		ctx, cancel := context.WithCancel(context.Background())
		log.Println("input " + url)

		w.cancel = cancel

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			w.logger.Debug("error after create request", err.Error())
			//w.crawler.ChangeStatus(false)
			output <- ""
			continue
		}

		resp, err := w.client.Do(req.WithContext(ctx))

		if err != nil {
			w.logger.Debug("error after request", err.Error())
			//w.crawler.ChangeStatus(false)
			output <- ""
			continue

		}

		if resp.StatusCode != http.StatusOK {
			w.logger.Error("err", "status code is not 200")
			output <- ""
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			w.logger.Debug("error after parse", err.Error())
			log.Println("error after parse " + err.Error())
			//w.crawler.ChangeStatus(false)
			output <- ""
			continue

		}
		output <- string(body)
		cancel()
	}
}

func (w *Worker) Close() {
	w.logger.Debug("msg", "worker closed")
	//w.cancel()
}

type Request struct {
	cancel context.CancelFunc
}
