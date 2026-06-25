import * as Setting from "../Setting";

const base = Setting.ServerUrl;
const lang = () => ({"Accept-Language": Setting.getAcceptLanguage()});
const jsonHeaders = () => ({"Content-Type": "application/json", ...lang()});

export function searchArtifactHub(q, page = 1) {
  return fetch(`${base}/api/search-artifact-hub?q=${encodeURIComponent(q)}&page=${page}&limit=20`, {
    credentials: "include", headers: lang(),
  }).then(r => r.json());
}

export function getHelmRepos() {
  return fetch(`${base}/api/get-helm-repos`, {credentials: "include", headers: lang()}).then(r => r.json());
}

export function addHelmRepo(repo) {
  return fetch(`${base}/api/add-helm-repo`, {
    method: "POST", credentials: "include", headers: jsonHeaders(), body: JSON.stringify(repo),
  }).then(r => r.json());
}

export function deleteHelmRepo(id) {
  return fetch(`${base}/api/delete-helm-repo?id=${id}`, {
    method: "POST", credentials: "include", headers: lang(),
  }).then(r => r.json());
}

export function getRepoCharts(url) {
  return fetch(`${base}/api/get-repo-charts?url=${encodeURIComponent(url)}`, {
    credentials: "include", headers: lang(),
  }).then(r => r.json());
}

export function getHelmChartValues(chart, repo, version) {
  return fetch(
    `${base}/api/get-helm-chart-values?chart=${encodeURIComponent(chart)}&repo=${encodeURIComponent(repo)}&version=${encodeURIComponent(version ?? "")}`,
    {credentials: "include", headers: lang()}
  ).then(r => r.json());
}

export function getHelmReleases(namespace = "all") {
  return fetch(`${base}/api/get-helm-releases?namespace=${namespace}`, {
    credentials: "include", headers: lang(),
  }).then(r => r.json());
}

export function installHelmChart(payload) {
  return fetch(`${base}/api/install-helm-chart`, {
    method: "POST", credentials: "include", headers: jsonHeaders(), body: JSON.stringify(payload),
  }).then(r => r.json());
}

export function upgradeHelmRelease(payload) {
  return fetch(`${base}/api/upgrade-helm-release`, {
    method: "POST", credentials: "include", headers: jsonHeaders(), body: JSON.stringify(payload),
  }).then(r => r.json());
}

export function rollbackHelmRelease(payload) {
  return fetch(`${base}/api/rollback-helm-release`, {
    method: "POST", credentials: "include", headers: jsonHeaders(), body: JSON.stringify(payload),
  }).then(r => r.json());
}

export function uninstallHelmRelease(payload) {
  return fetch(`${base}/api/uninstall-helm-release`, {
    method: "POST", credentials: "include", headers: jsonHeaders(), body: JSON.stringify(payload),
  }).then(r => r.json());
}

export function getHelmReleaseHistory(name, namespace) {
  return fetch(
    `${base}/api/get-helm-release-history?name=${encodeURIComponent(name)}&namespace=${encodeURIComponent(namespace)}`,
    {credentials: "include", headers: lang()}
  ).then(r => r.json());
}
