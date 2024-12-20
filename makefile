trigger:
	git commit -am'Updated pull' && git push
	git tag 0.0.3 --force
	goreleaser release --clean
	