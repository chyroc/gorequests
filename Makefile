coverage:
	./.github/test.sh
	go tool cover -html=coverage.txt
