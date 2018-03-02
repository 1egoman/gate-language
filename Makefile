.PHONY: build

namespace = "rgausnet"
image = "gate-language"
TAG ?= $(shell git rev-parse --short HEAD)

build:
	time docker build -t "$(namespace)/${image}:0" .

tag:
	docker tag "$(namespace)/$(image):0" "$(namespace)/$(image):$(TAG)"

publish:
	heroku container:push web -a lovelace-cloud

run:
	docker run -it \
		-v `pwd`/server:/go/src/app \
		-v `pwd`:/pwd \
		"$(namespace)/$(image):0" sh -c "go build && $(CMD)"

test:
	@make run CMD='go test $(ARGS)'

serve:
	@make run CMD='cp app /pwd/app'
	./app serve $(ARGS)
