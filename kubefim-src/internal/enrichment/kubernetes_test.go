package enrichment

import (
	"strings"
	"testing"
	"time"

	"kubefim/internal/event"
)

func TestSplitContainerID(t *testing.T) {
	id := strings.Repeat("b", 64)
	gotID, gotRuntime := splitContainerID("containerd://" + id)
	if gotID != id || gotRuntime != "containerd" {
		t.Fatalf("splitContainerID() = %q, %q", gotID, gotRuntime)
	}
	for _, invalid := range []string{"", id, "containerd://short", "containerd://" + strings.Repeat("z", 64)} {
		if gotID, gotRuntime := splitContainerID(invalid); gotID != "" || gotRuntime != "" {
			t.Fatalf("accepted invalid ID %q", invalid)
		}
	}
}

func TestUpdatePodIndexesAndDeletesContainer(t *testing.T) {
	id := strings.Repeat("c", 64)
	var value pod
	value.Metadata.Name = "api-123"
	value.Metadata.Namespace = "production"
	value.Metadata.UID = "pod-uid"
	value.Spec.NodeName = "worker-1"
	value.Status.ContainerStatuses = []containerStatus{{Name: "api", Image: "example/api:v1", ImageID: "sha256:image", ContainerID: "containerd://" + id}}

	cache := &Kubernetes{byID: make(map[string]event.Kubernetes), byPod: make(map[string][]string), misses: make(map[string]time.Time), updated: make(chan struct{})}
	cache.updatePod("ADDED", value)
	if got := cache.byID[id]; got.PodName != "api-123" || got.ContainerName != "api" {
		t.Fatalf("indexed metadata = %+v", got)
	}
	cache.updatePod("DELETED", value)
	if _, found := cache.byID[id]; found {
		t.Fatal("deleted pod remains indexed")
	}
}

func TestUpdatePodIgnoresUnknownWatchEvent(t *testing.T) {
	cache := &Kubernetes{byID: make(map[string]event.Kubernetes), byPod: make(map[string][]string), misses: make(map[string]time.Time), updated: make(chan struct{})}
	cache.updatePod("BOOKMARK", pod{})
	select {
	case <-cache.updated:
		t.Fatal("unknown watch event changed the cache")
	default:
	}
}
