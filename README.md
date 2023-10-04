### Crawler
Основная функция пакета - это обход url и получение данных. Конкурентность реализована двумя каналами на входе и выходе из основного тела crawler. На вход (```Input chan string```) отправляем url с которого необходимо получить данные, на выходе получаем структуру ```Output chan Response```, в которой указан url с которого производилась загрузка данных, и тело ответа (body) в байтах. 
## Примеры использования

# Инициализация crawler
crawler инициализируется путем вызова функции ```NewCrawler(ctx context.Context, concurrency int, client *http.Client)```, в которую передаем контекст (ctx), количество горутин которые будут скачивать и возвращать вам необходимые данные (concurrency), и http клиент (client). 
```
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

var transport *http.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).DialContext,
	}
var client http.Client = http.Client{Transport: transport, Timeout: time.Second * 10}

crawler1 := crawler.NewCrawler(ctx, 5, &client)
```
# Запрос данных
Необходимо реализовать отправку данных в канал и чтение данных из канала.
Отправка данных 
```
for _, url := range urls {
			crawler.InputProm <- crawler.Request{Url: url, Counter: 0}
		}
```
Получение данных
```
go func() {
	for Result := range crawler.Output {
		fmt.Printf("url: %v\n", Result.Url)
        fmt.Printf("body: %v\n", string(Result.Body))
    }
}()
```
Почему мы отправляем массив битов, а не строку? Для унификации пакеты было решено передавать массив байтов, поскольку остается возможно выполнить чтение данных в структуру (unmarshal).\n
Зачем возвращать Url и для чего нужен counter? на выходе можно реализовать проверку полученных данных и в случае неудачи выполнить повторный запрос. А что бы он не выполнялся бесконечное количество раз, его можно ограничить с помощью счетчика
```
go func() {
    for Result := range crawler.Output {
        if len(Result.Body) == 0 {
            if Result.UrlPrometheus.I > 3 {//устанавливаем предельное количество попыток получения данных
                logger.Debug("parsing target", "url", url)
                continue
            }
            logger.Debug("nill body", "url", url)
            crawler.Input <- Result
            continue
        }
        var result resultBody
        err := json.Unmarshal(Result.Body, &result)
        if err != nil {
            logger.Error("error parse json", "url", Result.Url, "error", err)
            continue
        }
        fmt.Printf("body: %#v\n", result)
    }
}()
```