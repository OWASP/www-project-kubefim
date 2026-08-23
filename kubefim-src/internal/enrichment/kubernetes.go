package enrichment

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"kubefim/internal/event"
)

const (
	serviceAccountToken = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCA    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

type Logger interface{ Printf(string, ...any) }

type Kubernetes struct {
	node      string
	apiURL    string
	token     string
	client    *http.Client
	proc      cgroupResolver
	logger    Logger
	mu        sync.RWMutex
	byID      map[string]event.Kubernetes
	byPod     map[string][]string
	byCgroup  map[containerKey]cachedContainer
	processes map[processKey]cachedContainer
	misses    map[string]time.Time
	updated   chan struct{}
}

type processKey struct {
	pid uint32
	containerKey
}

type containerKey struct {
	cgroupID       uint64
	mountNamespace uint32
	pidNamespace   uint32
}

type cachedContainer struct {
	container event.Container
	expires   time.Time
}

func NewInCluster(node, procRoot string, logger Logger) (*Kubernetes, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	if node == "" || host == "" || port == "" {
		return nil, errors.New("NODE_NAME and Kubernetes service environment are required")
	}
	token, err := os.ReadFile(serviceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}
	ca, err := os.ReadFile(serviceAccountCA)
	if err != nil {
		return nil, fmt.Errorf("read service account CA: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		return nil, errors.New("service account CA contains no certificates")
	}
	return &Kubernetes{
		node: node, apiURL: "https://" + net.JoinHostPort(host, port), token: strings.TrimSpace(string(token)),
		client: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots}}},
		proc:   newCgroupResolver(procRoot), logger: logger,
		byID: make(map[string]event.Kubernetes), byPod: make(map[string][]string),
		byCgroup:  make(map[containerKey]cachedContainer),
		processes: make(map[processKey]cachedContainer), misses: make(map[string]time.Time),
		updated: make(chan struct{}),
	}, nil
}

// Run maintains a node-filtered list/watch cache. It reconnects from a fresh
// list whenever the API server closes the watch or its resource version expires.
func (k *Kubernetes) Run(ctx context.Context) {
	for ctx.Err() == nil {
		resourceVersion, ok := k.refresh(ctx)
		if ok {
			k.watch(ctx, resourceVersion)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (k *Kubernetes) Enrich(ctx context.Context, value event.Event) event.Event {
	container, ok := k.resolveProcess(value)
	if !ok {
		return value
	}
	value.Container = container
	if container.ID == "" {
		return value
	}
	metadata, found, updated := k.lookup(container.ID)
	if !found && !k.beginWait(container.ID) {
		return value
	}
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	for !found {
		select {
		case <-ctx.Done():
			return value
		case <-timer.C:
			return value
		case <-updated:
			metadata, found, updated = k.lookup(container.ID)
		}
	}
	if found {
		value.Kubernetes = metadata
	}
	return value
}

func (k *Kubernetes) beginWait(containerID string) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	now := time.Now()
	if retry, found := k.misses[containerID]; found && now.Before(retry) {
		return false
	}
	k.misses[containerID] = now.Add(5 * time.Second)
	return true
}

func (k *Kubernetes) lookup(containerID string) (event.Kubernetes, bool, <-chan struct{}) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	metadata, found := k.byID[containerID]
	return metadata, found, k.updated
}

func (k *Kubernetes) notifyLocked() {
	close(k.updated)
	k.updated = make(chan struct{})
}

func (k *Kubernetes) resolveProcess(value event.Event) (event.Container, bool) {
	identity := containerKey{cgroupID: value.CgroupID, mountNamespace: value.MountNamespace, pidNamespace: value.PIDNamespace}
	key := processKey{pid: value.PID, containerKey: identity}
	now := time.Now()
	k.mu.RLock()
	cgroup, cgroupFound := k.byCgroup[identity]
	cached, found := k.processes[key]
	k.mu.RUnlock()
	if cgroupFound && now.Before(cgroup.expires) {
		return cgroup.container, true
	}
	if found && now.Before(cached.expires) {
		return cached.container, true
	}
	container, ok := k.proc.Resolve(value.PID)
	if !ok {
		return event.Container{}, false
	}
	k.mu.Lock()
	if len(k.processes) > 4096 {
		for candidate, value := range k.processes {
			if now.After(value.expires) {
				delete(k.processes, candidate)
			}
		}
	}
	k.processes[key] = cachedContainer{container: container, expires: now.Add(30 * time.Second)}
	if value.CgroupID != 0 {
		k.byCgroup[identity] = cachedContainer{container: container, expires: now.Add(time.Minute)}
	}
	if len(k.byCgroup) > 8192 {
		for id, value := range k.byCgroup {
			if now.After(value.expires) {
				delete(k.byCgroup, id)
			}
		}
	}
	k.mu.Unlock()
	return container, true
}

func (k *Kubernetes) refresh(ctx context.Context) (string, bool) {
	query := url.Values{"fieldSelector": {"spec.nodeName=" + k.node}}
	requestContext, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestContext, http.MethodGet, k.apiURL+"/api/v1/pods?"+query.Encode(), nil)
	if err != nil {
		k.logger.Printf("build Kubernetes pod request: %v", err)
		return "", false
	}
	req.Header.Set("Authorization", "Bearer "+k.token)
	resp, err := k.client.Do(req)
	if err != nil {
		k.logger.Printf("refresh Kubernetes pod cache: %v", err)
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		k.logger.Printf("refresh Kubernetes pod cache: status=%s body=%q", resp.Status, strings.TrimSpace(string(body)))
		return "", false
	}
	var list podList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		k.logger.Printf("decode Kubernetes pod list: %v", err)
		return "", false
	}
	next, nextByPod := make(map[string]event.Kubernetes), make(map[string][]string)
	for _, pod := range list.Items {
		for id, metadata := range podMetadata(pod) {
			next[id] = metadata
			nextByPod[pod.Metadata.UID] = append(nextByPod[pod.Metadata.UID], id)
		}
	}
	k.mu.Lock()
	k.byID, k.byPod = next, nextByPod
	for id := range next {
		delete(k.misses, id)
	}
	k.notifyLocked()
	k.mu.Unlock()
	return list.Metadata.ResourceVersion, true
}

func (k *Kubernetes) watch(ctx context.Context, resourceVersion string) {
	query := url.Values{
		"fieldSelector": {"spec.nodeName=" + k.node}, "resourceVersion": {resourceVersion},
		"watch": {"true"}, "timeoutSeconds": {"300"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.apiURL+"/api/v1/pods?"+query.Encode(), nil)
	if err != nil {
		k.logger.Printf("build Kubernetes pod watch: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+k.token)
	resp, err := k.client.Do(req)
	if err != nil {
		if ctx.Err() == nil {
			k.logger.Printf("watch Kubernetes pods: %v", err)
		}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		k.logger.Printf("watch Kubernetes pods: status=%s", resp.Status)
		return
	}
	decoder := json.NewDecoder(resp.Body)
	for {
		var update podWatchEvent
		if err := decoder.Decode(&update); err != nil {
			if ctx.Err() == nil && !errors.Is(err, io.EOF) {
				k.logger.Printf("decode Kubernetes pod watch: %v", err)
			}
			return
		}
		k.updatePod(update.Type, update.Object)
	}
}

func (k *Kubernetes) updatePod(updateType string, pod pod) {
	if updateType != "ADDED" && updateType != "MODIFIED" && updateType != "DELETED" {
		return
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, id := range k.byPod[pod.Metadata.UID] {
		delete(k.byID, id)
	}
	delete(k.byPod, pod.Metadata.UID)
	if updateType == "DELETED" {
		k.notifyLocked()
		return
	}
	for id, metadata := range podMetadata(pod) {
		k.byID[id] = metadata
		k.byPod[pod.Metadata.UID] = append(k.byPod[pod.Metadata.UID], id)
		delete(k.misses, id)
	}
	k.notifyLocked()
}

func podMetadata(pod pod) map[string]event.Kubernetes {
	result := make(map[string]event.Kubernetes)
	for _, status := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		id, _ := splitContainerID(status.ContainerID)
		if id != "" {
			result[id] = event.Kubernetes{Node: pod.Spec.NodeName, Namespace: pod.Metadata.Namespace, PodName: pod.Metadata.Name, PodUID: pod.Metadata.UID, ContainerName: status.Name, Image: status.Image, ImageID: status.ImageID}
		}
	}
	return result
}

func splitContainerID(value string) (string, string) {
	runtime, id, found := strings.Cut(value, "://")
	if !found || len(id) != 64 {
		return "", ""
	}
	for _, char := range id {
		if !strings.ContainsRune("0123456789abcdef", char) {
			return "", ""
		}
	}
	return id, runtime
}

type podList struct {
	Metadata struct {
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Items []pod `json:"items"`
}
type podWatchEvent struct {
	Type   string `json:"type"`
	Object pod    `json:"object"`
}
type pod struct {
	Metadata struct{ Name, Namespace, UID string } `json:"metadata"`
	Spec     struct {
		NodeName string `json:"nodeName"`
	} `json:"spec"`
	Status struct {
		ContainerStatuses     []containerStatus `json:"containerStatuses"`
		InitContainerStatuses []containerStatus `json:"initContainerStatuses"`
	} `json:"status"`
}
type containerStatus struct{ Name, Image, ImageID, ContainerID string }
