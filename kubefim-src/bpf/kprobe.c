//go:build ignore
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>

// A minimal event structure
struct event_t {
    __u32 pid;
    __u32 uid;
    char comm[16];
    char path[256];
};

// Use the stable perf event array for communication
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
} events SEC(".maps");


SEC("tracepoint/syscalls/sys_enter_openat")
int tracepoint_openat(struct trace_event_raw_sys_enter *ctx) {
    struct event_t event = {};

    event.pid = bpf_get_current_pid_tgid() >> 32;
    event.uid = bpf_get_current_uid_gid();
    bpf_get_current_comm(&event.comm, sizeof(event.comm));
    
    // Read the path argument from the syscall
    const char *path_ptr = (const char *)ctx->args[1];
    bpf_probe_read_user_str(&event.path, sizeof(event.path), path_ptr);

    // Submit the event
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &event, sizeof(event));
    
    return 0;
}

char __license[] SEC("license") = "Dual MIT/GPL";