package crawler

import (
	"context"
	"io"
	"log"
	"net/http"

	logging "github.com/griddis/go-logger"
)

type Crawler struct {
	Status      bool
	Logger      logging.Logger
	concurrency int
	Workers     []*Worker
	Input       chan Request
	Output      chan Response
}

type Request struct {
	Url     string
	Counter int
}

type Response struct {
	Url     string
	Counter int
	Body    []byte
}

func NewCrawler(ctx context.Context, concurrency int, client *http.Client) *Crawler {
	logger := logging.FromContext(ctx)
	craw := Crawler{
		Status:      true,
		Logger:      logger,
		concurrency: concurrency,
		Workers:     make([]*Worker, concurrency),
		Input:       make(chan Request),
		Output:      make(chan Response),
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
	logger  logging.Logger
	client  *http.Client
	cancel  context.CancelFunc
	crawler *Crawler
	status  bool
}

func NewWorker(logger logging.Logger, client *http.Client, craw *Crawler) *Worker {
	return &Worker{
		logger:  logger,
		status:  true,
		crawler: craw,
		client:  client,
	}
}

func (w *Worker) Run(input chan Request, output chan Response) {
	w.logger.Debug("status", "start")

	for request := range input {
		request.Counter++
		ctx, cancel := context.WithCancel(context.Background())
		w.logger.Debug("input " + request.Url)

		w.cancel = cancel

		req, err := http.NewRequest(http.MethodGet, request.Url, nil)
		if err != nil {
			w.logger.Debug("error after create request", err.Error())
			//w.crawler.ChangeStatus(false)
			output <- Response{Url: request.Url, Counter: request.Counter, Body: nil}
			continue
		}

		resp, err := w.client.Do(req.WithContext(ctx))

		if err != nil {
			w.logger.Debug("error after request", err.Error())
			//w.crawler.ChangeStatus(false)
			output <- Response{Url: request.Url, Counter: request.Counter, Body: nil}
			continue

		}

		if resp.StatusCode != http.StatusOK {
			w.logger.Error("err", "status code is not 200")
			output <- Response{Url: request.Url, Counter: request.Counter, Body: nil}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			w.logger.Debug("error after parse", err.Error())
			log.Println("error after parse " + err.Error())
			//w.crawler.ChangeStatus(false)
			output <- Response{Url: request.Url, Counter: request.Counter, Body: nil}
			continue

		}
		output <- Response{Url: request.Url, Counter: request.Counter, Body: body}
		cancel()
	}
}

func (w *Worker) Close() {
	w.logger.Debug("msg", "worker closed")
	//w.cancel()
}

// type RequestCanselCtx struct {
// 	cancel context.CancelFunc
// }
