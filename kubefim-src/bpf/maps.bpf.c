//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include "src/common.h"

struct events_map_t {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, 256);
};

struct events_map_t events SEC(".maps");

struct pending_events_map_t {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 16384);
    __type(key, __u64);
    __type(value, struct event_t);
};

struct pending_events_map_t pending_events SEC(".maps");

struct scratch_event_map_t {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct event_t);
};

struct scratch_event_map_t scratch_event SEC(".maps");

char LICENSE[] SEC("license") = "Dual MIT/GPL";
