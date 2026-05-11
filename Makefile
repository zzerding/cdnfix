.PHONY: docker-build docker push docker-run-push docker-run-test

docker-build:
	docker build -t zzerding/cdn .
docker-push:
	docker push zzerding/cdn
docker-run-push:
	docker run --rm --env-file=.env -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdn -u https://www.zuohaomc.com/join/ push
docker-run-query:
	docker run --rm  --env-file=.env  -v $(PWD)/.task_push.cache:/root/.task_push.cache -v $(PWD)/.task_refresh.cache:/root/.task_refresh.cache zzerding/cdn  query