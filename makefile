helm-docs:
	helm-docs


VERSION := "0.0.7"
trigger:

	git commit --allow-empty -am'Updated pull' && git push
	git tag ${VERSION} --force
	DOCKER_CONFIG=$$HOME/.docker/planesailingio goreleaser release --clean
	$(MAKE) publish-chart

publish-chart:
	cd /lake/git/charts/charts && \
		helm package --version "${VERSION}" --app-version "${VERSION}" /lake/git/goTFHub/helm/gotfhub && \
		cd .. && \
		helm repo index . --url https://charts.planesailing.io && \
		git add . && git commit -m'add new chart' && git push
