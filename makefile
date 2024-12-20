trigger:
	git commit -am'Updated pull' && git push
	git tag 0.0.3 --force
	DOCKER_CONFIG=$$HOME/.docker/planesailingio goreleaser release --clean
