fmt:
	@sh scripts/fmt.sh

test:
	@sh scripts/test.sh v4

upgrade-go-micro:
	@sh scripts/upgrade-go-micro.sh v4 latest