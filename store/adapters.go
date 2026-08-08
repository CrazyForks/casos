package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// helmChartAdapter encodes app-aware install value patches keyed by chart name.
// It makes an app usable right after installation (e.g. a reachable service
// type) without requiring the user to know chart internals.
type helmChartAdapter struct {
	valuesPatches map[string]interface{}
	// generatedValuesPatchFn mints per-install values such as a random secret.
	// The values preview skips it so the install dialog stays reproducible.
	generatedValuesPatchFn func(explicitValues map[string]interface{}) (map[string]interface{}, error)
	// clusterHostBindings publish the cluster addresses into a chart's host
	// allowlist. The values preview skips them: they depend on cluster state.
	clusterHostBindings []clusterHostBinding
	// preservedValuePaths are carried over from the installed release on upgrade.
	preservedValuePaths [][]string
}

// clusterHostBinding names where a chart keeps the hostnames its app answers
// to. CasOS publishes app store installs as NodePort services, so users reach
// them through a cluster node address the chart cannot know; apps that check
// the Host header reject those requests outright (Nextcloud's trusted_domains
// with a 400, Django's ALLOWED_HOSTS, Grafana's server.domain with broken
// links). A binding is data, not code: supporting another app means naming a
// values path and a format.
type clusterHostBinding struct {
	// path is where the assembled host list is written.
	path []string
	// sourcePaths hold hosts the app must keep answering to, in priority
	// order — typically its configured domain, then the destination's current
	// value, then route hostnames. Each is read from the install values and
	// from the chart defaults, and may hold a string or a list of strings.
	sourcePaths [][]string
	// render turns the host list into the shape the chart expects.
	render func(hosts []string) interface{}
}

// clusterContext carries install-time facts the values preview cannot know.
// nodeIPs is lazy so charts without a binding never trigger the node list call.
type clusterContext struct {
	nodeIPs     func() []string
	releaseName string
	namespace   string
}

// helmChartAdapterRegistry maps canonical chart names to install adaptations.
// Explicit user values always win over adapter patches.
var helmChartAdapterRegistry = map[string]helmChartAdapter{
	"grafana":  {valuesPatches: nodePortServiceValuesPatch()},
	"pgadmin4": {valuesPatches: nodePortServiceValuesPatch()},
	"n8n":      {valuesPatches: nodePortServiceValuesPatch()},
	"superset": {
		valuesPatches:          supersetValuesPatches(),
		generatedValuesPatchFn: supersetSecretKeyPatch,
		preservedValuePaths:    supersetPreservedValuePaths,
	},
	"nextcloud": {
		valuesPatches:       nodePortServiceValuesPatch(),
		clusterHostBindings: nextcloudTrustedDomainsBinding,
	},
}

// nextcloudTrustedDomainsBinding keeps Nextcloud reachable through the node
// address the NodePort patch above publishes: any Host outside trusted_domains
// gets a 400 "Access through untrusted domain". The chart templates this path
// into NEXTCLOUD_TRUSTED_DOMAINS, and doing so drops its own default of
// nextcloud.host — the Host header its probes send to /status.php — so that
// name heads the source list.
var nextcloudTrustedDomainsBinding = []clusterHostBinding{{
	path: []string{"nextcloud", "trustedDomains"},
	sourcePaths: [][]string{
		{"nextcloud", "host"},
		{"nextcloud", "trustedDomains"},
		{"httpRoute", "hostnames"},
	},
	render: helmHostStringList,
}}

// nodePortServiceValuesPatch returns a fresh patch each call so the registry
// never shares mutable state.
func nodePortServiceValuesPatch() map[string]interface{} {
	return map[string]interface{}{
		"service": map[string]interface{}{"type": "NodePort"},
	}
}

// supersetValuesPatches also clears the chart's node port default: it ships
// `http: nil`, which YAML reads as the string "nil" and renders as an invalid
// `nodePort: nil`. A real null lets Kubernetes allocate a free port, where a
// fixed number would collide with the next install.
func supersetValuesPatches() map[string]interface{} {
	patches := nodePortServiceValuesPatch()
	service := patches["service"].(map[string]interface{})
	service["nodePort"] = map[string]interface{}{"http": nil}
	patches["bootstrapScript"] = supersetBootstrapScript
	return patches
}

const (
	supersetDriverTarget    = "/tmp/pgdrivers"
	supersetPsycopg2Version = "2.9.10"
	supersetSecretKeyEnvVar = "SUPERSET_SECRET_KEY"
)

// supersetBootstrapScript makes the PostgreSQL driver importable at pod start:
// the apache/superset image does not ship psycopg2 and the app venv has no pip,
// so the system python installs it into a writable target dir. The import probe
// keeps images that already carry it, and every restart on an offline cluster,
// from reaching for PyPI at all; the short timeout fails fast when it must.
// ${PYTHONPATH:+...} drops the separator when PYTHONPATH is unset, where a
// trailing colon would put the working directory on sys.path.
var supersetBootstrapScript = strings.Join([]string{
	"#!/bin/bash",
	"if ! python -c 'import psycopg2' >/dev/null 2>&1; then",
	fmt.Sprintf(
		"  /usr/local/bin/python3 -m pip install --no-cache-dir --disable-pip-version-check --timeout 15 --retries 1 --target %s 'psycopg2-binary==%s' || true",
		supersetDriverTarget, supersetPsycopg2Version,
	),
	"fi",
	fmt.Sprintf("export PYTHONPATH=%s${PYTHONPATH:+:${PYTHONPATH}}", supersetDriverTarget),
}, "\n")

// supersetPreservedValuePaths must survive an upgrade: Superset encrypts stored
// database credentials with SECRET_KEY, so a new key orphans them. The second
// path is where releases installed before the switch keep their key.
var supersetPreservedValuePaths = [][]string{
	{"extraSecretEnv", supersetSecretKeyEnvVar},
	{"configOverrides", "secret"},
}

// supersetSecretKeyPatch generates a random SECRET_KEY per install, because
// Superset refuses to start with the packaged default. extraSecretEnv reaches
// superset/config.py through the env Secret, avoiding configOverrides, whose
// merge order the chart documents as unspecified.
func supersetSecretKeyPatch(explicitValues map[string]interface{}) (map[string]interface{}, error) {
	if supersetSecretKeyAlreadySet(explicitValues) {
		return nil, nil
	}
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("generate Superset SECRET_KEY: %w", err)
	}
	return map[string]interface{}{
		"extraSecretEnv": map[string]interface{}{supersetSecretKeyEnvVar: hex.EncodeToString(buf)},
	}, nil
}

// supersetSecretKeyAlreadySet reports whether configOverrides already defines a
// SECRET_KEY. superset_config.py is read after the environment, so it wins and
// a generated key would only look like a rotation on upgrade.
func supersetSecretKeyAlreadySet(values map[string]interface{}) bool {
	overrides, ok := values["configOverrides"].(map[string]interface{})
	if !ok {
		return false
	}
	secret, ok := overrides["secret"].(string)
	return ok && strings.TrimSpace(secret) != ""
}

// applyHelmChartAdapter merges chart-specific compatibility values into the
// install values; user-set top-level keys are left untouched.
func applyHelmChartAdapter(ch *chart.Chart, values, explicitValues map[string]interface{}, includeDynamic bool, cluster clusterContext) error {
	if ch == nil {
		return nil
	}
	adapter, ok := helmChartAdapterRegistry[helmChartAdapterKey(ch)]
	if !ok {
		return nil
	}
	patches := adapter.valuesPatches
	if includeDynamic && adapter.generatedValuesPatchFn != nil {
		generated, err := adapter.generatedValuesPatchFn(explicitValues)
		if err != nil {
			return fmt.Errorf("apply Helm chart adapter for %s: %w", ch.Name(), err)
		}
		patches = mergedHelmAdapterPatches(patches, generated)
	}
	for topKey, patch := range patches {
		if adapterPatchExplicitlyOverridden(explicitValues, topKey, patch) {
			continue
		}
		if err := mergeHelmValueOverrides(values, map[string]interface{}{topKey: patch}, nil); err != nil {
			return fmt.Errorf("apply Helm chart adapter for %s: %w", ch.Name(), err)
		}
	}
	if includeDynamic {
		if err := applyClusterHostBindings(ch, values, adapter.clusterHostBindings, cluster); err != nil {
			return fmt.Errorf("apply Helm chart adapter for %s: %w", ch.Name(), err)
		}
	}
	return nil
}

// applyClusterHostBindings writes the reachable host list into each bound
// values path. Unlike the patches above this never yields to explicit values:
// the list is additive, so the user's own hosts head it and CasOS only appends
// the addresses the app would otherwise turn away.
func applyClusterHostBindings(ch *chart.Chart, values map[string]interface{}, bindings []clusterHostBinding, cluster clusterContext) error {
	for _, binding := range bindings {
		if len(binding.path) == 0 || binding.render == nil {
			continue
		}
		hosts := clusterReachableHosts(ch, values, binding.sourcePaths, cluster)
		if len(hosts) == 0 {
			continue
		}
		if err := mergeHelmValueOverrides(values, helmValuePatch(binding.path, binding.render(hosts)), nil); err != nil {
			return err
		}
	}
	return nil
}

// clusterReachableHosts assembles every address the app is reached through:
// the binding's own sources first, then loopback, the in-cluster service names
// and the node addresses. Order is stable and duplicates are dropped, so a
// binding that renders only the first host still gets the app's own domain.
func clusterReachableHosts(ch *chart.Chart, values map[string]interface{}, sourcePaths [][]string, cluster clusterContext) []string {
	seen := map[string]bool{}
	var hosts []string
	add := func(host string) {
		host = strings.TrimSpace(host)
		if host == "" || seen[host] {
			return
		}
		seen[host] = true
		hosts = append(hosts, host)
	}
	for _, path := range sourcePaths {
		for _, host := range helmValueHosts(values, path) {
			add(host)
		}
		if ch != nil {
			for _, host := range helmValueHosts(ch.Values, path) {
				add(host)
			}
		}
	}
	add("localhost")
	for _, host := range cluster.inClusterHosts(ch, values) {
		add(host)
	}
	if cluster.nodeIPs != nil {
		for _, ip := range cluster.nodeIPs() {
			add(ip)
		}
	}
	return hosts
}

// inClusterHosts returns the service DNS names other pods reach the release
// through; sidecars such as the Nextcloud metrics exporter scrape the app that
// way, and a host allowlist that omits them turns those callers away.
func (cluster clusterContext) inClusterHosts(ch *chart.Chart, values map[string]interface{}) []string {
	if cluster.releaseName == "" || cluster.namespace == "" {
		return nil
	}
	names := cluster.releaseServiceNames(ch, values)
	hosts := make([]string, 0, len(names))
	for _, name := range names {
		hosts = append(hosts, fmt.Sprintf("%s.%s.svc.cluster.local", name, cluster.namespace))
	}
	return hosts
}

// releaseServiceNames lists the names a release's service may carry. It follows
// the fullname template nearly every chart copies from `helm create`:
// fullnameOverride decides outright, otherwise the release name is used when it
// already carries the chart name (or its nameOverride) and is prefixed to it
// when it does not. The bare release name heads the list either way, for the
// charts that name a service after the release alone — a host allowlist
// tolerates a spare entry, not a missing one.
func (cluster clusterContext) releaseServiceNames(ch *chart.Chart, values map[string]interface{}) []string {
	names := []string{cluster.releaseName}
	if fullname := helmNameOverride(ch, values, "fullnameOverride"); fullname != "" {
		names = append(names, fullname)
	} else {
		name := helmNameOverride(ch, values, "nameOverride")
		if name == "" && ch != nil {
			name = helmChartAdapterKey(ch)
		}
		if name != "" && !strings.Contains(cluster.releaseName, name) {
			names = append(names, cluster.releaseName+"-"+name)
		}
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(names))
	for _, name := range names {
		name = helmDNSName(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}

// helmNameOverride reads a top-level naming override the way Helm coalesces it:
// the install values win over the chart's own default.
func helmNameOverride(ch *chart.Chart, values map[string]interface{}, key string) string {
	sources := []map[string]interface{}{values}
	if ch != nil {
		sources = append(sources, ch.Values)
	}
	for _, source := range sources {
		if override, ok := helmValueAtPath(source, []string{key}).(string); ok {
			if override = strings.TrimSpace(override); override != "" {
				return override
			}
		}
	}
	return ""
}

// helmDNSName applies the `trunc 63 | trimSuffix "-"` every fullname template
// ends with, so a long release name yields the name Kubernetes actually holds.
func helmDNSName(name string) string {
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.TrimSuffix(name, "-")
}

// helmHostStringList renders the host list as a YAML list of strings.
func helmHostStringList(hosts []string) interface{} {
	items := make([]interface{}, 0, len(hosts))
	for _, host := range hosts {
		items = append(items, host)
	}
	return items
}

// helmValueHosts reads a values path holding one host or a list of hosts,
// ignoring entries that are not strings.
func helmValueHosts(values map[string]interface{}, path []string) []string {
	switch typed := helmValueAtPath(values, path).(type) {
	case string:
		return []string{typed}
	case []string:
		return append([]string(nil), typed...)
	case []interface{}:
		hosts := make([]string, 0, len(typed))
		for _, item := range typed {
			if host, ok := item.(string); ok {
				hosts = append(hosts, host)
			}
		}
		return hosts
	}
	return nil
}

// helmValueAtPath returns the leaf a path points at, or nil if any parent is
// missing or is not a map.
func helmValueAtPath(values map[string]interface{}, path []string) interface{} {
	for index, key := range path {
		if values == nil {
			return nil
		}
		if index == len(path)-1 {
			return values[key]
		}
		next, ok := values[key].(map[string]interface{})
		if !ok {
			return nil
		}
		values = next
	}
	return nil
}

// helmValuePatch wraps a leaf value in the nested maps its path describes.
func helmValuePatch(path []string, value interface{}) map[string]interface{} {
	patch := value
	for index := len(path) - 1; index > 0; index-- {
		patch = map[string]interface{}{path[index]: patch}
	}
	return map[string]interface{}{path[0]: patch}
}

// preserveHelmChartAdapterValues copies the installed release's adapter-owned
// secrets into the upgrade values, where the adapter then reads them as
// explicit input and leaves them alone. Caller-set values win.
func preserveHelmChartAdapterValues(actionConfig *action.Configuration, ch *chart.Chart, releaseName string, values map[string]interface{}) {
	if ch == nil || actionConfig == nil || actionConfig.Releases == nil {
		return
	}
	adapter, ok := helmChartAdapterRegistry[helmChartAdapterKey(ch)]
	if !ok || len(adapter.preservedValuePaths) == 0 {
		return
	}
	installedRelease, err := actionConfig.Releases.Last(releaseName)
	if err != nil || installedRelease == nil || installedRelease.Chart == nil {
		return
	}
	installed, err := chartutil.CoalesceValues(installedRelease.Chart, cloneHelmValues(installedRelease.Config))
	if err != nil {
		return
	}
	for _, path := range adapter.preservedValuePaths {
		copyHelmValueIfUnset(installed, values, path)
	}
}

// copyHelmValueIfUnset copies one non-empty string leaf between value trees,
// creating parent maps on the way and never replacing an existing leaf.
func copyHelmValueIfUnset(source, target map[string]interface{}, path []string) {
	if len(path) == 0 {
		return
	}
	parents, leaf := path[:len(path)-1], path[len(path)-1]
	for _, key := range parents {
		next, ok := source[key].(map[string]interface{})
		if !ok {
			return
		}
		source = next
	}
	value, ok := source[leaf].(string)
	if !ok || value == "" {
		return
	}
	for _, key := range parents {
		switch existing := target[key].(type) {
		case map[string]interface{}:
			target = existing
		case nil:
			next := map[string]interface{}{}
			target[key] = next
			target = next
		default:
			return
		}
	}
	if _, exists := target[leaf]; exists {
		return
	}
	target[leaf] = value
}

// mergedHelmAdapterPatches overlays generated patches on the static ones
// without mutating the registry entry.
func mergedHelmAdapterPatches(static, generated map[string]interface{}) map[string]interface{} {
	if len(generated) == 0 {
		return static
	}
	merged := make(map[string]interface{}, len(static)+len(generated))
	for topKey, patch := range static {
		merged[topKey] = patch
	}
	for topKey, patch := range generated {
		merged[topKey] = patch
	}
	return merged
}

func helmChartAdapterKey(ch *chart.Chart) string {
	return strings.ToLower(strings.TrimSpace(ch.Name()))
}

// adapterPatchExplicitlyOverridden reports whether the user explicitly set
// any leaf of the patch within explicitValues. Leaf-level checking (not the
// top-level key) keeps the adapter active when the user only touched a
// sibling key, e.g. service.port while the patch targets service.type.
func adapterPatchExplicitlyOverridden(explicitValues map[string]interface{}, topKey string, patch interface{}) bool {
	explicit, exists := explicitValues[topKey]
	if !exists {
		return false
	}
	patchMap, patchIsMap := patch.(map[string]interface{})
	explicitMap, explicitIsMap := explicit.(map[string]interface{})
	if !patchIsMap || !explicitIsMap {
		return true
	}
	for key, subPatch := range patchMap {
		if adapterPatchExplicitlyOverridden(explicitMap, key, subPatch) {
			return true
		}
	}
	return false
}
