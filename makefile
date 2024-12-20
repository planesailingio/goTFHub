helm-docs:
	helm-docs


VERSION := "0.0.5"
trigger:

	git commit -am'Updated pull' && git push
	git tag ${VERSION} --force
	DOCKER_CONFIG=$$HOME/.docker/planesailingio goreleaser release --clean
	$(MAKE) publish-chart

publish-chart:
	cd /lake/git/charts && \
		helm package --version "${VERSION}" --app-version "${VERSION}" /lake/git/goTFHub/helm/gotfhub && \
		helm repo index . && \
		git add . && git commit -m'add new chart' && git push
