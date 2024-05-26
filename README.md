# tencent cloud cdn refresh tool
* refresh cdn cache by url and url dir
* push cache by url
* query refresh status
  

## TODO
* add aliyun cloud cdn 


## how to use
### create scretId and scretKey in https://console.cloud.tencent.com/cam/capi
### set enviroment variables or edit .env
  * copy env.example to.env.
  * .env is the default enviroment file, you cat use -e to set enviroment file
```
export SECRET_ID=xxxxxx
export SECRET_KEY=xxxxxx
```

### use source run commad
1. `git clone https://github.com/zzerding/refresh-cdn.git`
2. cd refresh-cdn
3. go run main.go
### use binary run commad
1. go install  github.com/zzerding/refresh-cdn
2. refresh-cdn 
### use docker run commad
1. docker run -rm -v $PWD:/app -e SECRET_ID=xxx -e SECRET_KEY=xxx zzerding/cnd 
2. docker run --rm --env-file=.env -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdn -u https://www.xxxx.com/join/ push
3. docker run --rm  --env-file=.env  -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdn  query

## example
#### set .env or set enviroment variables

* refresh cdn cache by url 
  
  `refresh-cdn -u https://www.xxxx.com refresh`

* refresh cdn cache by dir
  `refresh-cdn -u https://www.xxxx.com/ refresh`

* query refresh status
  `refresh-cdn query`

* refresh cdn cache by url file
  ```
  echo https://www.xxx.com/s?wd=tencent > /tmp/test.txt
  echo https://www.xxxx.com/ > /tmp/test.html
  refresh-cdn -f /tmp/test.txt refresh
  ```

* use othev env file 
  `refresh-cdn -e /tmp/.env.shanghai -f /tmp/urlList.txt refresh`

* use debug mode
  `refresh-cdn -d query`

