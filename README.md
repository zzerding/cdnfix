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
1. `git clone https://github.com/zzerding/cdnfix.git`
2. cd cdnfix
3. go run main.go
### use binary run commad
1. go install  github.com/zzerding/cdnfix
2. cdnfix 
### use docker run commad
1.use -e args
```
 docker run -rm -v $PWD:/app -e SECRET_ID=xxx -e SECRET_KEY=xxx zzerding/refresh-cnd 
```
2. save cache to local
```
 docker run --rm --env-file=.env -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdnfix -u https://www.xxxx.com/join/ push
```
3. query status
```
 docker run --rm  --env-file=.env  -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdnfix  query
```

## example
#### set .env or set enviroment variables

* refresh cdn cache by url 
  
  `cdnfix -u https://www.xxxx.com refresh`

* refresh cdn cache by dir
  `cdnfix -u https://www.xxxx.com/ refresh`

* query refresh status
  `cdnfix query`

* refresh cdn cache by url file
  ```
  echo https://www.xxx.com/s?wd=tencent > /tmp/test.txt
  echo https://www.xxxx.com/ > /tmp/test.html
  cdnfix -f /tmp/test.txt refresh
  ```

* use othev env file 
  `cdnfix -e /tmp/.env.shanghai -f /tmp/urlList.txt refresh`

* use debug mode
  `cdnfix -d query`

